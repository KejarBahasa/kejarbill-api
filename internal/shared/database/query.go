package database

import (
	"strconv"
	"strings"
)

func BuildBulkInsertQuery(tableName string, columns []string, rowCount int) string {
	if rowCount <= 0 {
		return ""
	}

	var sb strings.Builder

	colCount := len(columns)

	sb.WriteString("INSERT INTO ")
	sb.WriteString(tableName)
	sb.WriteString(" (")
	sb.WriteString(strings.Join(columns, ", "))
	sb.WriteString(") VALUES ")

	for i := 0; i < rowCount; i++ {
		sb.WriteString("(")

		for j := 0; j < colCount; j++ {
			placeholderNum := i*colCount + j + 1

			sb.WriteString("$" + strconv.Itoa(placeholderNum))

			if j < colCount-1 {
				sb.WriteString(", ")
			}
		}

		sb.WriteString(")")

		if i < rowCount-1 {
			sb.WriteString(", ")
		}
	}

	return sb.String()
}
