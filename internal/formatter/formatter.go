package formatter

import (
	"encoding/csv"
	"fmt"
	"io"
	"strconv"
	"strings"

	"3mfanalyzer/internal/parser"
)

func FormatAsText(data *parser.Parser3MF, writer io.Writer) error {
	fmt.Fprintf(writer, "3MF File Analysis\n")
	fmt.Fprintf(writer, "=================\n\n")

	if len(data.Plates) == 0 {
		fmt.Fprintf(writer, "No plates found in the file.\n")
		return nil
	}

	for _, plate := range data.Plates {
		fmt.Fprintf(writer, "Plate %d: %s\n", plate.PlateID, plate.PlateName)

		if len(plate.Objects) == 0 {
			fmt.Fprintf(writer, "  No objects on this plate.\n")
			continue
		}

		groups := parser.GroupObjectsByName(plate.Objects)
		for _, group := range groups {
			if group.Type == "assembly" {
				fmt.Fprintf(writer, "  %d x %s (assembly)\n", group.Count, group.Name)
			} else {
				fmt.Fprintf(writer, "  %d x %s\n", group.Count, group.Name)
			}

			if group.Type == "assembly" && len(group.Components) > 0 {
				fmt.Fprintf(writer, "    Components:\n")
				for _, comp := range group.Components {
					fmt.Fprintf(writer, "      - %s (ID: %d, Source: %s)\n", comp.Name, comp.ID, comp.SourceFile)
				}
			}
		}
		fmt.Fprintf(writer, "\n")
	}

	return nil
}

func FormatAsCSV(data *parser.Parser3MF, writer io.Writer) error {
	csvWriter := csv.NewWriter(writer)
	defer csvWriter.Flush()

	headers := []string{
		"PlateID", "PlateName", "ObjectName", "ObjectType", "Count",
		"ComponentCount", "ComponentNames", "ComponentFiles",
	}

	if err := csvWriter.Write(headers); err != nil {
		return fmt.Errorf("failed to write CSV headers: %w", err)
	}

	for _, plate := range data.Plates {
		if len(plate.Objects) == 0 {
			record := []string{
				strconv.Itoa(plate.PlateID),
				plate.PlateName,
				"", "", "0",
				"0", "", "",
			}
			if err := csvWriter.Write(record); err != nil {
				return fmt.Errorf("failed to write CSV record: %w", err)
			}
			continue
		}

		groups := parser.GroupObjectsByName(plate.Objects)
		for _, group := range groups {
			var componentNames, componentFiles []string
			for _, comp := range group.Components {
				componentNames = append(componentNames, comp.Name)
				componentFiles = append(componentFiles, comp.SourceFile)
			}

			record := []string{
				strconv.Itoa(plate.PlateID),
				plate.PlateName,
				group.Name,
				group.Type,
				strconv.Itoa(group.Count),
				strconv.Itoa(len(group.Components)),
				strings.Join(componentNames, ";"),
				strings.Join(componentFiles, ";"),
			}

			if err := csvWriter.Write(record); err != nil {
				return fmt.Errorf("failed to write CSV record: %w", err)
			}
		}
	}

	return nil
}

