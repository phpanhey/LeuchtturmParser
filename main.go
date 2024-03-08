package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/go-rod/rod"
)

func main() {

	html := extractHtml("https://zum-pusdorper-leuchtturm.jimdosite.com/mittach-rouladen-taxi/")

	// use goquery to query the html
	imageUrl := extractImageUrl(html)

	imageName := "menue.jpg"

	dowloadImage(imageUrl, imageName)

	menueText := getTextFromImage(imageName)

	modifiedMenueText := deleteFirstLineWithWord(menueText, "Montag")

	currentMenue := extractCurrentMenue(modifiedMenueText)

	fmt.Println(currentMenue)

}

func extractImageUrl(html string) string {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	// check error
	if err != nil {
		log.Fatal(err)
	}
	// find first image tag
	res := doc.Find("img").First().AttrOr("srcset", "")
	return strings.Split(res, " ")[0]
}

func extractHtml(url string) string {
	browser := rod.New().MustConnect()

	page := browser.MustPage(url)

	page.MustWaitLoad()

	html := page.MustHTML()
	return html
}

func dowloadImage(url string, imagenName string) {
	// download the image

	res, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()

	// create a file to store the image
	file, err := os.Create(imagenName)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	_, err = io.Copy(file, res.Body)
	if err != nil {
		log.Fatal(err)
	}
}

func getTextFromImage(imageName string) string {
	// Setup the tesseract command
	cmd := exec.Command("tesseract", imageName, "stdout")

	// Capture both standard output and standard error
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Execute the command
	err := cmd.Run()

	// Check for errors
	if err != nil {
		log.Fatal(err)
	}

	return stdout.String()

}

func extractCurrentMenue(menueText string) string {
	// german weekday names
	weekdaysDe := []string{"Montag", "Dienstag", "Mittwoch", "Donnerstag", "Freitag"}
	currentWeekday := getCurrentWeekday(weekdaysDe)
	fmt.Println(currentWeekday)

	// iterate over every line of text when weekday found append next 4 lines to result
	// finally return result
	lines := strings.Split(menueText, "\n")
	var result string
	for i := 0; i < len(lines); i++ {
		if strings.Contains(lines[i], currentWeekday) {
			for j := 1; j < 5; j++ {
				result += lines[i+j] + "\n"
			}
		}
	}
	return strings.ReplaceAll(result, "\n", " ")
}

func getCurrentWeekday(weekdaysDe []string) string {
	weekdaysEn := []string{"Monday", "Tuesday", "Wednesday", "Thursday", "Friday"}

	for i := 0; i < len(weekdaysEn); i++ {
		if weekdaysEn[i] == time.Now().Weekday().String() {
			return weekdaysDe[i]
		}
	}
	return "No weekday found"
}

func deleteFirstLineWithWord(s, word string) string {
	lines := strings.Split(s, "\n")
	for i, line := range lines {
		if strings.Contains(line, word) {
			lines = append(lines[:i], lines[i+1:]...)
			break
		}
	}
	return strings.Join(lines, "\n")
}
