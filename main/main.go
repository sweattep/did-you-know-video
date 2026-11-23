
package main

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

type Summary struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Extract     string `json:"extract"`

	Thumbnail struct {
		Source string `json:"source"`
	} `json:"thumbnail"`
}

func main() {
	mainPage := get(
		"https://en.wikipedia.org/wiki/Main_Page",
	)

	re := regexp.MustCompile(
		`href="/wiki/([^":#]+)"`,
	)

	m := re.FindStringSubmatch(
		extractDYK(mainPage),
	)

	if len(m) < 2 {
		panic("no article found")
	}

	article := m[1]

	var summary Summary

	body := get(
		"https://en.wikipedia.org/api/rest_v1/page/summary/" +
			article,
	)

	json.Unmarshal(
		[]byte(body),
		&summary,
	)

	script := buildScript(summary)

	os.WriteFile(
		"script.txt",
		[]byte(script),
		0644,
	)

	img := download(summary.Thumbnail.Source)

	os.WriteFile(
		"image.jpg",
		img,
		0644,
	)

	exec.Command(
		"espeak-ng",
		"-f",
		"script.txt",
		"-w",
		"narration.wav",
	).Run()

	exec.Command(
		"ffmpeg",
		"-y",
		"-loop", "1",
		"-i", "image.jpg",
		"-i", "narration.wav",
		"-c:v", "libx264",
		"-c:a", "aac",
		"-pix_fmt", "yuv420p",
		"-shortest",
		"output.mp4",
	).Run()
}

func buildScript(s Summary) string {
	var b strings.Builder

	b.WriteString("Did you know?\n\n")
	b.WriteString(s.Extract)
	b.WriteString("\n\n")
	b.WriteString("The subject is ")
	b.WriteString(s.Title)
	b.WriteString(".\n\n")
	b.WriteString(s.Description)
	b.WriteString("\n\n")
	b.WriteString("For more information, see the full Wikipedia article.\n")

	return b.String()
}

func extractDYK(html string) string {
	start := strings.Index(
		html,
		`id="mp-dyk"`,
	)

	if start == -1 {
		panic("dyk not found")
	}

	html = html[start:]

	end := strings.Index(
		html,
		"</ul>",
	)

	return html[:end]
}

func get(url string) string {
	resp, _ := http.Get(url)
	defer resp.Body.Close()

	b, _ := io.ReadAll(resp.Body)

	return string(b)
}

func download(url string) []byte {
	resp, _ := http.Get(url)
	defer resp.Body.Close()

	var buf bytes.Buffer
	io.Copy(&buf, resp.Body)

	return buf.Bytes()
}
