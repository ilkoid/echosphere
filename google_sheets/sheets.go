package google_sheets

import (
	"context"
	"echosphere/apis"
	"fmt"
	"os"

	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

func PrepareDataForSheets(reviews []apis.ReviewCondensed) [][]interface{} {
	values := [][]interface{}{
		{"Id", "Text"},
	}
	for _, review := range reviews {
		values = append(values, []interface{}{review.ID, review.Text})
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

	_, err = srv.Spreadsheets.Values.Append(spreadsheetId, writeRange, valueRange).
		ValueInputOption("USER_ENTERED").
		Do()
	if err != nil {
		return fmt.Errorf("unable to write data to sheet: %v", err)
	}

	return nil
}
