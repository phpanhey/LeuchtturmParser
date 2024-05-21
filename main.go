package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
)

func main() {

	html := extractHtml("https://zum-pusdorper-leuchtturm.jimdosite.com/mittach-rouladen-taxi/")

	// use goquery to query the html
	imageUrl := extractImageUrl(html)

	imageName := "menue.jpg"

	dowloadImage(imageUrl, imageName)

	menueText := getTextFromImage(imageName)

	modifiedMenueText := deleteFirstLineWithWord(menueText, "Montag")

	menueJson := extractMenueAsJson(modifiedMenueText)

	saveFile("menue.json", menueJson)

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
	u := launcher.New().Set("no-sandbox").MustLaunch()
	browser := rod.New().ControlURL(u).MustConnect()

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

func extractMenueAsJson(menueText string) map[string]string {
	res := make(map[string]string)
	sweekdaysDe := []string{"Montag", "Dienstag", "Mittwoch", "Donnerstag", "Freitag"}
	sweekdaysEn := []string{"Monday", "Tuesday", "Wednesday", "Thursday", "Friday"}
	for i := 0; i < len(sweekdaysDe); i++ {
		res[sweekdaysEn[i]] = extractMenueForDay(menueText, sweekdaysDe[i])
	}
	return res
}

func extractMenueForDay(menueText string, weekddayDe string) string {
	lines := strings.Split(menueText, "\n")
	var result string
	for i := 0; i < len(lines); i++ {
		if strings.Contains(lines[i], weekddayDe) {
			for j := 1; j < 4; j++ {
				result += lines[i+j] + "\n"
			}
		}
	}
	if strings.Contains(result, "geschlossen") {
		return "geschlossen"
	}
	return strings.ReplaceAll(cleanString(result), "\n", " ")

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

func saveFile(fileName string, menueJson map[string]string) {
	jsonData, err := json.Marshal(menueJson)
	if err != nil {
		fmt.Println("Error marshalling JSON:", err)
		return
	}
	err = os.WriteFile(fileName, jsonData, 0644)
	if err != nil {
		fmt.Println("Error writing JSON to file:", err)
		return
	}
}

func cleanString(s string) string {
	re := regexp.MustCompile(`\s+\d+(?:,\d+)*,[a-z]\s*`)
	return re.ReplaceAllString(s, " ")
}
