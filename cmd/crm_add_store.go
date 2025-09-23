package cmd

import (
	"fmt"
	"os"
	"time"

	"farmix-cli/internal/bitrix"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	addStoreDealID  string
	addStoreStoreID string
	addStoreDryRun  bool
	addStoreCurrency string
)

var crmAddStoreCmd = &cobra.Command{
	Use:   "crm-add-store",
	Short: "Создание документа прихода на склад из товаров сделки Bitrix24",
	Long: `Создание документа прихода на склад с товарами из сделки Bitrix24.

Команда выполнит следующие действия:
1. Получит информацию о сделке и товарах из Bitrix24
2. Проверит, включен ли складской учет
3. Получит и проверит информацию о складе (название, статус активности)
4. Создаст документ прихода (тип документа 'S' - оприходование)
5. Добавит все товары из сделки в документ с количествами из сделки
6. Проведет документ для обновления складских остатков

Документ будет использовать ID товаров для точности и будет автоматически проведен.

Используйте флаг --dry-run для предварительного просмотра без внесения изменений.`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := runCRMAddStore(); err != nil {
			fmt.Fprintf(os.Stderr, "Ошибка: %v\n", err)
			os.Exit(1)
		}
	},
}

func runCRMAddStore() error {
	// Validate parameters
	if err := bitrix.ValidateDealID(addStoreDealID); err != nil {
		return fmt.Errorf("неверный ID сделки: %v", err)
	}

	// Get webhook URL from config
	webhookURL := viper.GetString("bitrix_webhook_url")
	if webhookURL == "" {
		return fmt.Errorf("bitrix_webhook_url не настроен. Пожалуйста, установите его в конфигурации ~/.farmix-cli")
	}

	// Use store_id from config if not specified via flag
	if addStoreStoreID == "1" { // Check if it's the default value
		configStoreID := viper.GetString("store_id")
		if configStoreID != "" {
			addStoreStoreID = configStoreID
		}
	}

	if addStoreDryRun {
		fmt.Printf("[ТЕСТОВЫЙ РЕЖИМ] Обработка сделки %s для создания документа прихода...\n", addStoreDealID)
	} else {
		fmt.Printf("Обработка сделки %s для создания документа прихода...\n", addStoreDealID)
	}

	// Create Bitrix24 client
	client := bitrix.NewClient(webhookURL)

	// Check if warehouse management is enabled
	fmt.Println("Проверка статуса складского учета...")
	enabled, err := client.CheckStoreDocumentMode()
	if err != nil {
		return fmt.Errorf("не удалось проверить статус складского учета: %v", err)
	}
	if !enabled {
		return fmt.Errorf("складской учет не включен в Bitrix24")
	}
	fmt.Println("Складской учет включен ✓")

	// Test API access to stores
	fmt.Println("Тестирование доступа к API складов...")
	stores, listErr := client.ListStores()
	if listErr != nil {
		return fmt.Errorf("нет доступа к API складов: %v\n\nПроверьте права доступа:\n1. Войдите в Bitrix24 → Разработчикам → Другое → Входящий вебхук\n2. Найдите ваш вебхук и нажмите \"Изменить\"\n3. Убедитесь, что включены права доступа:\n   - catalog (Торговый каталог)\n   - crm (CRM)\n4. Сохраните изменения и попробуйте снова\n\nТакже проверьте:\n- Складской учет активирован в Bitrix24 (Настройки → Настройки модулей → Торговый каталог)\n- Созданы склады в разделе \"Магазин\" → \"Склады\"", listErr)
	}

	if len(stores) == 0 {
		return fmt.Errorf("список складов пуст\n\nВозможные причины:\n1. Склады не созданы в Bitrix24 (Магазин → Склады)\n2. Недостаточно прав доступа к складам (проверьте права 'catalog' в вебхуке)\n3. Складской учет отключен")
	}

	fmt.Printf("Найдено %d складов ✓\n", len(stores))

	// Get store information
	fmt.Printf("Получение информации о складе ID %s...\n", addStoreStoreID)
	store, err := client.GetStore(addStoreStoreID)
	if err != nil {
		return fmt.Errorf("не удалось получить информацию о складе: %v", err)
	}

	// Check if store was found (empty fields indicate not found)
	if store.ID == 0 && store.Title == "" {
		fmt.Printf("Склад с ID %s не найден. Получение списка доступных складов...\n", addStoreStoreID)
		stores, listErr := client.ListStores()
		if listErr != nil {
			return fmt.Errorf("склад ID %s не найден и не удалось получить список складов: %v", addStoreStoreID, listErr)
		}

		fmt.Printf("Доступные склады:\n")
		for _, s := range stores {
			status := "неактивен"
			if s.Active == "Y" {
				status = "активен"
			}
			fmt.Printf("  - ID: %d, Название: %s, Статус: %s\n", s.ID, s.Title, status)
		}

		return fmt.Errorf("склад с ID %s не найден. Используйте один из доступных складов", addStoreStoreID)
	}

	if store.Active != "Y" {
		return fmt.Errorf("склад ID %s не активен (статус: '%s', название: '%s')", addStoreStoreID, store.Active, store.Title)
	}

	fmt.Printf("Склад: %s (ID: %d) ✓\n", store.Title, store.ID)

	// Get deal information
	fmt.Println("Получение информации о сделке...")
	deal, err := client.GetDealWithAmount(addStoreDealID)
	if err != nil {
		return fmt.Errorf("не удалось получить информацию о сделке: %v", err)
	}
	fmt.Printf("Сделка: %s\n", deal.Title)

	// Get products from deal
	fmt.Println("Получение товаров из сделки...")
	products, err := client.GetExistingProductRows(addStoreDealID)
	if err != nil {
		return fmt.Errorf("не удалось получить товары из сделки: %v", err)
	}

	if len(products) == 0 {
		return fmt.Errorf("в сделке %s не найдено товаров", addStoreDealID)
	}

	fmt.Printf("Найдено %d товаров в сделке\n", len(products))

	if addStoreDryRun {
		fmt.Printf("[ТЕСТОВЫЙ РЕЖИМ] Товары, которые будут добавлены на склад:\n")
		for _, product := range products {
			fmt.Printf("  - ID товара: %s, Количество: %.2f, Цена: %.2f %s\n",
				product.ProductID.String(), product.Quantity, product.Price, addStoreCurrency)
		}
		fmt.Printf("[ТЕСТОВЫЙ РЕЖИМ] Будет создан документ прихода с %d товарами\n", len(products))
		fmt.Printf("[ТЕСТОВЫЙ РЕЖИМ] Товары будут добавлены на склад: %s (ID: %s)\n", store.Title, store.ID)
		return nil
	}

	// Create warehouse receipt document
	fmt.Println("Создание документа прихода...")
	documentID, err := client.CreateStoreDocument(deal, addStoreCurrency, fmt.Sprintf("Приход товаров по сделке %s", addStoreDealID))
	if err != nil {
		return fmt.Errorf("не удалось создать документ прихода: %v", err)
	}
	fmt.Printf("Создан документ с ID: %s\n", documentID)

	// Document created with status "N" (not confirmed) to allow adding elements

	// Add products to document
	fmt.Println("Добавление товаров в документ...")
	fmt.Printf("Добавляем товары в документ ID: %s на склад ID: %s\n", documentID, addStoreStoreID)
	err = client.AddElementsToStoreDocument(documentID, products, addStoreStoreID)
	if err != nil {
		return fmt.Errorf("не удалось добавить товары в документ: %v", err)
	}
	fmt.Printf("Добавлено %d товаров в документ\n", len(products))

	// Confirm document
	fmt.Println("Проведение документа...")
	err = client.ConfirmStoreDocument(documentID)
	if err != nil {
		return fmt.Errorf("не удалось провести документ: %v", err)
	}

	fmt.Printf("Успешно создан и проведен документ прихода %s\n", documentID)
	fmt.Printf("Товары добавлены на склад (ID склада: %s):\n", addStoreStoreID)
	for _, product := range products {
		fmt.Printf("  - ID товара: %s, Количество: %.2f\n", product.ProductID.String(), product.Quantity)
	}

	return nil
}

func init() {
	crmAddStoreCmd.Flags().StringVar(&addStoreDealID, "deal-id", "", "ID сделки Bitrix24 (обязательно)")
	crmAddStoreCmd.Flags().StringVar(&addStoreStoreID, "store-id", "1", "ID склада (по умолчанию: 1, или из конфигурации ~/.farmix-cli)")
	crmAddStoreCmd.Flags().StringVar(&addStoreCurrency, "currency", "RUB", "Валюта для документа (по умолчанию: RUB)")
	crmAddStoreCmd.Flags().BoolVar(&addStoreDryRun, "dry-run", false, "Предварительный просмотр без внесения изменений")

	crmAddStoreCmd.MarkFlagRequired("deal-id")

	rootCmd.AddCommand(crmAddStoreCmd)
}