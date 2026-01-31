package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

func main() {

	html, err := getHtml("https://zum-pusdorper-leuchtturm.jimdosite.com/mittach-rouladen-taxi/")

	if err != nil {
		log.Fatal(err)
	}

	imageUrl := extractImageUrl(html)

	imageName := "menue.jpg"

	dowloadImage(imageUrl, imageName)

	menueText := getTextFromImage(imageName)

	modifiedMenueText := deleteFirstLineWithWord(menueText, "Montag")

	menueJson := extractMenueAsJson(modifiedMenueText)

	printMenue(menueJson)

	//saveFile("menue.json", menueJson)

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

func getHtml(url string) (string, error) {
	client := &http.Client{
		Timeout: 15 * time.Second,
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}

	// Common Chrome desktop UA
	req.Header.Set("User-Agent",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) "+
			"AppleWebKit/537.36 (KHTML, like Gecko) "+
			"Chrome/120.0.0.0 Safari/537.36")

	// Optional but helpful extra headers
	req.Header.Set("Accept",
		"text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	req.Header.Set("Connection", "keep-alive")

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

func dowloadImage(url string, imagenName string) {
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
	cmd := exec.Command(
		"tesseract",
		imageName,
		"stdout",
		"-l", "deu",
	)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()

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

/*func saveFile(fileName string, menueJson map[string]string) {
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
}*/

func cleanString(s string) string {
	re := regexp.MustCompile(`\s+\d+(?:,\d+)*,[a-z]\s*`)
	return re.ReplaceAllString(s, " ")
}

func printMenue(menueJson map[string]string) {
	weekday := time.Now().Weekday()
	fmt.Println(menueJson[string(weekday)])
}
