package cmd

import (
	"fmt"
	"os"

	"farmix-cli/internal/bitrix"
	"farmix-cli/internal/formatter"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	reportFormat string
)

var crmReportCmd = &cobra.Command{
	Use:   "crm-report",
	Short: "Вывести отчет по активным сделкам Bitrix24",
	Long: `Вывести таблицу по всем сделкам, которые не в финальном состоянии.

Команда выполнит следующие действия:
1. Получит список сделок из Bitrix24
2. Отфильтрует сделки, исключив финальные статусы (по умолчанию WON и LOST)
3. Отсортирует сделки по ID (возрастание)
4. Выведет таблицу с полями:
   - ID сделки
   - Название сделки
   - Дата создания
   - Рассчетная стоимость м/ч (кастомное поле)
   - Рассчетная стоимость ч/ч (кастомное поле)
   - Рассчетная стоимость материала (кастомное поле)
   - Итого стоимость изготовления (кастомное поле)
   - Оплата получена (кастомное поле)

Для работы команды необходимо настроить коды кастомных полей в ~/.farmix-cli:

report_custom_fields:
  machine_cost: "UF_CRM_XXXXX"
  human_cost: "UF_CRM_XXXXX"
  material_cost: "UF_CRM_XXXXX"
  total_cost: "UF_CRM_XXXXX"
  payment_received: "UF_CRM_XXXXX"

Коды полей можно найти в Bitrix24: CRM -> Настройки -> Поля -> Сделки`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := runCRMReport(); err != nil {
			fmt.Fprintf(os.Stderr, "Ошибка: %v\n", err)
			os.Exit(1)
		}
	},
}

func runCRMReport() error {
	// Get webhook URL from config
	webhookURL := viper.GetString("bitrix_webhook_url")
	if webhookURL == "" {
		return fmt.Errorf("bitrix_webhook_url не настроен. Пожалуйста, установите его в конфигурации ~/.farmix-cli")
	}

	// Get custom fields configuration
	customFields := bitrix.ReportCustomFields{
		MachineCost:     viper.GetString("report_custom_fields.machine_cost"),
		HumanCost:       viper.GetString("report_custom_fields.human_cost"),
		MaterialCost:    viper.GetString("report_custom_fields.material_cost"),
		TotalCost:       viper.GetString("report_custom_fields.total_cost"),
		PaymentReceived: viper.GetString("report_custom_fields.payment_received"),
	}

	// Validate that at least some custom fields are configured
	if customFields.MachineCost == "" && customFields.HumanCost == "" &&
		customFields.MaterialCost == "" && customFields.TotalCost == "" &&
		customFields.PaymentReceived == "" {
		return fmt.Errorf("не настроены коды кастомных полей в ~/.farmix-cli\n\n" +
			"Добавьте в конфигурационный файл секцию:\n\n" +
			"report_custom_fields:\n" +
			"  machine_cost: \"UF_CRM_XXXXX\"\n" +
			"  human_cost: \"UF_CRM_XXXXX\"\n" +
			"  material_cost: \"UF_CRM_XXXXX\"\n" +
			"  total_cost: \"UF_CRM_XXXXX\"\n" +
			"  payment_received: \"UF_CRM_XXXXX\"\n\n" +
			"Коды полей можно найти в Bitrix24: CRM -> Настройки -> Поля -> Сделки")
	}

	// Get excluded statuses from config (default to WON and LOST)
	excludedStatuses := viper.GetStringSlice("report_excluded_statuses")
	if len(excludedStatuses) == 0 {
		excludedStatuses = []string{"WON", "LOST"}
	}

	fmt.Println("Получение списка сделок из Bitrix24...")

	// Create Bitrix24 client
	client := bitrix.NewClient(webhookURL)

	// Get deals with custom fields
	deals, err := client.ListDealsWithCustomFields(customFields, excludedStatuses)
	if err != nil {
		return fmt.Errorf("не удалось получить список сделок: %v", err)
	}

	if len(deals) == 0 {
		fmt.Println("Нет активных сделок для отображения")
		return nil
	}

	fmt.Printf("Найдено %d активных сделок\n\n", len(deals))

	// Format and output report
	switch reportFormat {
	case "csv":
		if err := formatter.FormatReportAsCSV(deals, os.Stdout); err != nil {
			return fmt.Errorf("не удалось сформировать CSV отчет: %v", err)
		}
	case "text":
		if err := formatter.FormatReportAsTable(deals, os.Stdout); err != nil {
			return fmt.Errorf("не удалось сформировать текстовый отчет: %v", err)
		}
	default:
		return fmt.Errorf("неподдерживаемый формат вывода: %s (поддерживаются: text, csv)", reportFormat)
	}

	return nil
}

func init() {
	crmReportCmd.Flags().StringVarP(&reportFormat, "format", "f", "text", "Формат вывода (text, csv)")

	rootCmd.AddCommand(crmReportCmd)
}
