package main

import (
	"bytes"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"hash"
	"io"
	"log"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/cavaliercoder/grab"
	ui "github.com/gizak/termui"
	"github.com/mholt/archiver"
	"github.com/olekukonko/tablewriter"
)

const (
	viewGoVersions = "goversions"
	viewGoFiles    = "gofiles"
)

var mutex = &sync.Mutex{}

var viewInFocus = viewGoFiles
var selectedGoVersion = 0
var selectedGoFile = 0

var goVersions []string
var goFiles []string
var doc *goquery.Document
var downloadFileLinks = make(map[string]string)
var downloadFileSHA = make(map[string]string)

var uiVersionsList = ui.NewList()
var uiFilesList = ui.NewList()

//var uiInfo = ui.NewPar("Info")
var uiInfo = ui.NewList()
var uiDownloadProgress = ui.NewGauge()
var uiDownloadSpeed = ui.NewSparklines()

func getGolangData(url string) {
	var err error
	doc, err = goquery.NewDocument(url)
	if err != nil {
		log.Fatal(err)
	}

	r, err := regexp.Compile("([a-zA-Z0-9.]+)")
	if err != nil {
		log.Fatal(err)
	}

	goVersions = nil
	doc.Find("div .expanded > h2").Each(func(i int, s *goquery.Selection) {
		goVersions = append(goVersions, r.FindString(s.Text()))
	})
}

func updateGoVersions() {
	var strs []string

	checkLimits()

	for index, s := range goVersions {
		if index == selectedGoVersion {
			strs = append(strs, fmt.Sprintf("[%s](fg-white,bg-green)", s))
		} else {
			strs = append(strs, fmt.Sprintf("%s", s))
		}
	}
	uiVersionsList.Items = strs
}

func addLoggingText(text string) {
	uiInfo.Items = append(uiInfo.Items, text)
	height := uiInfo.GetHeight()
	startPosition := len(uiInfo.Items) - height + 2
	if startPosition < 0 {
		startPosition = 0
	}
	uiInfo.Items = uiInfo.Items[startPosition:]

}

func downloadFile(url string) error {

	addLoggingText(fmt.Sprintf("Downloading %s...\n", url))
	respch, err := grab.GetAsync(".", url)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error downloading %s: %v\n", url, err)
		os.Exit(1)
	}

	addLoggingText(fmt.Sprintf("Initializing download...\n"))

	uiDownloadProgress.Percent = 0

	var resp *grab.Response

	go func() {
		t := time.NewTicker(100 * time.Millisecond)
		var sinceLastTime uint64
		var spdata []int
		for {
			select {
			case r := <-respch:
				if r != nil {
					resp = r
					addLoggingText(fmt.Sprintf("Downloading...\n"))
				}
			case <-t.C:
				if resp == nil {
					continue
				}
				if resp.Error != nil {
					fmt.Fprintf(os.Stderr, "Error downloading %s: %v\n", url, resp.Error)
					//return resp.Error
					return
				}
				if resp.IsComplete() {
					addLoggingText(fmt.Sprintf("Successfully downloaded to ./%s\n", resp.Filename))

					uiDownloadProgress.Percent = 100

					addLoggingText("Verifying...\n")
					sha := downloadFileSHA[getGoFile()]

					var hasher hash.Hash
					if len(sha) == 64 {
						hasher = sha256.New()
					} else if len(sha) == 40 {
						hasher = sha1.New()
					} else {
						addLoggingText(fmt.Sprintf("Unknown hash length of %d.\n", len(sha)))
						return
					}

					f, err := os.Open(resp.Filename)
					if err != nil {
						addLoggingText(fmt.Sprintf(err.Error()))
						return
					}
					defer f.Close()

					if _, err := io.Copy(hasher, f); err != nil {
						addLoggingText(fmt.Sprintf(err.Error()))
						return
					}
					fileSHA := hex.EncodeToString(hasher.Sum(nil))

					if sha != fileSHA {
						addLoggingText("Download file doesn't match SHA.\n")
						addLoggingText(fmt.Sprintf("File SHA: %s\nExpected SHA:\n%s\n", fileSHA, sha))
						return
					}

					addLoggingText(fmt.Sprintf("File SHA matches: %s\n", fileSHA))

					addLoggingText(fmt.Sprintf("Extracting...\n"))
					if strings.Contains(resp.Filename, ".zip") {
						archiver.Zip.Open(resp.Filename, ".")
					} else if strings.Contains(resp.Filename, ".tar.gz") {
						archiver.TarGz.Open(resp.Filename, ".")
					}
					os.Rename("go", getGoVersion())
					addLoggingText(fmt.Sprintf("Done extracting.\n"))
					t.Stop()
					return
				}

				uiDownloadProgress.Percent = int(100 * resp.Progress())

				thisTime := resp.BytesTransferred() - sinceLastTime
				sinceLastTime = resp.BytesTransferred()
				spdata = append(spdata, int(thisTime))
				uiDownloadSpeed.Lines[0].Data = spdata
			}
		}
	}()
	return nil
}

func updateFilesView(s string) error {
	var err error

	var b bytes.Buffer
	table := tablewriter.NewWriter(&b)

	var headers []string
	var data [][]string
	goFiles = nil
	doc.Find("div[id=\"" + strings.TrimSpace(s) + "\"] > .expanded > table > thead > tr > th").Each(func(i int, s *goquery.Selection) {
		headers = append(headers, s.Text())
	})

	doc.Find("div[id=\"" + strings.TrimSpace(s) + "\"] > .expanded > table > tbody > tr").Each(func(i int, tr *goquery.Selection) {
		var row []string
		var a *goquery.Selection
		tr.Find("td").Each(func(i int, td *goquery.Selection) {

			if a == nil {
				a = td.Find("a")
			}
			if link, exists := a.Attr("href"); exists {
				downloadFileLinks[a.Text()] = link
			}

			tt := td.Find("tt")
			if tt != nil && a != nil {
				downloadFileSHA[a.Text()] = tt.Text()
			}
			row = append(row, td.Text())
		})
		if len(row) > 0 {
			goFiles = append(goFiles, row[0])
		}
		data = append(data, row)
	})
	table.SetHeader(headers)
	//table.SetBorders(tablewriter.Border{Left: true, Top: false, Right: true, Bottom: false})
	table.SetBorder(false)
	//table.SetCenterSeparator("|")
	table.AppendBulk(data)
	table.Render()

	checkLimits()

	strs := strings.Split(b.String(), "\n")

	for index, s := range strs {
		if index == selectedGoFile+2 { // +2 To skip headers.
			strs[index] = "[" + strings.TrimRight(s, "\n") + "](fg-white,bg-blue)\n"
			break
		}
	}

	uiFilesList.Items = strs

	return err
}

func getGoVersion() string {
	return goVersions[selectedGoVersion]
}

func getGoFile() string {
	return goFiles[selectedGoFile]
}

func checkLimits() {
	if selectedGoVersion < 0 {
		selectedGoVersion = 0
	} else if selectedGoVersion >= len(goVersions) {
		selectedGoVersion = len(goVersions) - 1
	}

	if selectedGoFile < 0 {
		selectedGoFile = 0
	} else if selectedGoFile >= len(goFiles) {
		selectedGoFile = len(goFiles) - 1
	}
}

func main() {

	url := "https://golang.org/dl/"
	log.Printf("Getting data from %s...\n", url)
	getGolangData(url)

	if err := ui.Init(); err != nil {
		panic(err)
	}
	defer ui.Close()

	spark := ui.NewSparkline()
	spark.Height = 8
	spark.LineColor = ui.ColorCyan
	spark.TitleColor = ui.ColorWhite

	uiDownloadSpeed.Add(spark)
	uiDownloadSpeed.Height = 11
	uiDownloadSpeed.BorderLabel = "Download Speed"

	uiFilesList.Height = len(goVersions) + 2
	uiFilesList.BorderLabel = "Go Files"
	uiFilesList.Width = 20
	uiFilesList.Border = true

	uiInfo.BorderLabel = "Info Panel"
	uiInfo.Items = []string{"Ready."}
	//uiInfo.Text = "Ready.\n"
	uiInfo.Height = 14

	updateFilesView(getGoVersion())

	uiDownloadProgress.LabelAlign = ui.AlignCenter
	uiDownloadProgress.Height = 3
	uiDownloadProgress.Border = true
	uiDownloadProgress.BorderLabel = "Download Progress"
	uiDownloadProgress.BarColor = ui.ColorRed

	updateGoVersions()
	uiVersionsList.BorderLabel = "Go Versions"
	uiVersionsList.Height = len(goVersions) + 2
	uiVersionsList.Width = 10

	// build layout
	ui.Body.AddRows(
		ui.NewRow(
			ui.NewCol(2, 0, uiVersionsList),
			ui.NewCol(10, 0, uiFilesList)),
		ui.NewRow(
			ui.NewCol(6, 0, uiDownloadProgress, uiDownloadSpeed),
			ui.NewCol(6, 0, uiInfo)))

	// calculate layout
	ui.Body.Align()

	ui.Merge("timer", ui.NewTimerCh(10*time.Millisecond))
	ui.Render(ui.Body)

	ui.Handle("/sys/kbd/q", func(ui.Event) {
		ui.StopLoop()
	})
	ui.Handle("/sys/kbd/Q", func(ui.Event) {
		ui.StopLoop()
	})
	ui.Handle("/sys/kbd/C-c", func(ui.Event) {
		ui.StopLoop()
	})
	ui.Handle("/sys/kbd/<up>", func(ui.Event) {
		mutex.Lock()
		defer mutex.Unlock()
		if viewInFocus == viewGoFiles {
			selectedGoFile--
			updateFilesView(getGoVersion())
		} else if viewInFocus == viewGoVersions {
			selectedGoVersion--
			updateGoVersions()
			updateFilesView(getGoVersion())
		}
	})
	ui.Handle("/sys/kbd/<enter>", func(ui.Event) {
		downloadFile(downloadFileLinks[getGoFile()])
	})
	ui.Handle("/sys/kbd/<down>", func(ui.Event) {
		mutex.Lock()
		defer mutex.Unlock()
		if viewInFocus == viewGoFiles {
			selectedGoFile++
			checkLimits()
			updateFilesView(getGoVersion())
		} else if viewInFocus == viewGoVersions {
			selectedGoVersion++
			checkLimits()
			updateGoVersions()
			updateFilesView(getGoVersion())
		}
	})
	ui.Handle("/sys/kbd/<tab>", func(ui.Event) {
		if viewInFocus == viewGoFiles {
			viewInFocus = viewGoVersions
			uiFilesList.BorderLabelBg = ui.ColorDefault
			//uiFilesList.BorderLabelFg = ui.ColorGreen
			uiVersionsList.BorderLabelBg = ui.ColorWhite
		} else if viewInFocus == viewGoVersions {
			viewInFocus = viewGoFiles
			uiFilesList.BorderLabelBg = ui.ColorWhite
			//uiFilesList.BorderLabelFg = ui.ColorBlack
			//uiVersionsList.BorderLabelFg = ui.ColorGreen
			uiVersionsList.BorderLabelBg = ui.ColorDefault
		}
	})
	ui.Handle("/timer/", func(e ui.Event) {
		ui.Render(ui.Body)
	})

	ui.Handle("/sys/wnd/resize", func(e ui.Event) {
		ui.Body.Width = ui.TermWidth()
		ui.Body.Align()
		ui.Clear()
		ui.Render(ui.Body)
	})

	uiFilesList.BorderLabelBg = ui.ColorWhite
	ui.Loop()
}
