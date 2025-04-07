package google_sheets

import (
	"context"
	processer "echosphere/processer"
	"fmt"
	"os"
	"reflect"

	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

func PrepareDataForSheets(reviews []processer.ReviewProcessed) [][]interface{} {
	if len(reviews) == 0 {
		return [][]interface{}{}
	}

	first := reviews[0]
	structFields := reflect.VisibleFields(reflect.TypeOf(first))
	structFieldNames := []string{}
	for _, field := range structFields {
		if field.Type.Kind() != reflect.Array && field.Type.Kind() != reflect.Struct {
			structFieldNames = append(structFieldNames, field.Name)
		}
	}

	interfacedFieldNames := make([]interface{}, len(structFieldNames))
	for i, v := range structFieldNames {
		interfacedFieldNames[i] = v
	}

	values := [][]interface{}{}
	values = append(values, interfacedFieldNames)

	for _, review := range reviews {
		interfacedFieldsValues := make([]interface{}, len(structFieldNames))
		currentFields := []reflect.Value{}
		for i, fieldName := range structFieldNames {
			interfacedFieldsValues[i] = fmt.Sprintf("%v", reflect.ValueOf(review).FieldByName(fieldName))
			currentFields = append(currentFields, reflect.ValueOf(review).FieldByName(fieldName))
		}
		values = append(values, interfacedFieldsValues)
	}
	return values
}

func WriteReviewsToSheet(values [][]interface{}, spreadsheetId string, writeRange string, credentialsFile string) error {
	ctx := context.Background()
	b, err := os.ReadFile(credentialsFile)
	if err != nil {
		return fmt.Errorf("unable to read service account key file: %v", err)
	}

	config, err := google.JWTConfigFromJSON(b, sheets.SpreadsheetsScope)
	if err != nil {
		return fmt.Errorf("unable to parse service account key to config: %v", err)
	}

	client := config.Client(ctx)

	srv, err := sheets.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return fmt.Errorf("unable to retrieve Sheets client: %v", err)
	}

	valueRange := &sheets.ValueRange{
		Values: values,
	}

	// first of all clear what was before
	clearRequest := &sheets.ClearValuesRequest{}
	_, err = srv.Spreadsheets.Values.Clear(spreadsheetId, "Лист1!A1:Z10000", clearRequest).Do()
	if err != nil {
		return fmt.Errorf("unable to clear sheet: %v", err)
	}

	// now write new data
	_, err = srv.Spreadsheets.Values.Append(spreadsheetId, writeRange, valueRange).
		ValueInputOption("RAW").
		Do()
	if err != nil {
		return fmt.Errorf("unable to write data to sheet: %v", err)
	}

	return nil
}
