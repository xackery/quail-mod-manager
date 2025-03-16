package qmm

import (
	"archive/zip"
	"bytes"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"quail-mod-manager/component"
	"quail-mod-manager/dialog"
	"quail-mod-manager/grass"
	"quail-mod-manager/handler"
	"quail-mod-manager/mw"
	"strings"

	"quail-mod-manager/ico"

	"github.com/xackery/quail/quail"
	"github.com/xackery/quail/wce"
	"github.com/xackery/wlk/cpl"
	"github.com/xackery/wlk/walk"
	"gopkg.in/yaml.v3"
)

var instance *Qmm

type Qmm struct {
	Entries []*QmmYamlEntry
}

type QmmYaml struct {
	Entry []QmmYamlEntry `yaml:"qmm"`
}

type QmmYamlEntry struct {
	IsEnabled    bool
	Name         string `yaml:"name"`
	ID           string `yaml:"id"`
	Version      string `yaml:"version"`
	QuailVersion string `yaml:"quail"`
	Image        string `yaml:"image"`
	ImageData    []byte
	imageBitmap  *walk.Bitmap
	URL          string `yaml:"url"`
	Author       string `yaml:"author"`
	Description  string `yaml:"description"`
}

func New() (*Qmm, error) {
	q := &Qmm{}
	err := q.loadConfig()
	if err != nil {
		dialog.ShowMessageBox("Error", err.Error()+"\nReverting to default settings", true)
		q = &Qmm{}
	}

	for _, e := range q.Entries {
		if len(e.ImageData) > 0 {
			e.imageBitmap = ico.Generate(e.ID, e.ImageData)
		}
	}

	instance = q
	handler.ImportModZipSubscribe(onModZip)
	handler.RemoveModSubscribe(onRemoveMod)
	handler.ImportModURLSubscribe(onModURL)
	handler.GenerateModSubscribe(onGenerateMod)
	handler.EnableModSubscribe(onEnableMod)

	rebuildModlist()
	return q, nil
}

func onModZip() {
	path, err := dialog.ShowOpen("Select a mod zip file", "Mod Zip Files (*.zip)", ".")
	if err != nil {
		if err.Error() == "show open: cancelled" {
			return
		}
		dialog.ShowMessageBox("Error", err.Error(), true)
		return
	}
	if path == "" {
		return
	}
	err = AddModZip(path)
	if err != nil {
		dialog.ShowMessageBox("Error", err.Error(), true)
		return
	}
}

func AddModZip(path string) error {
	q := instance
	if q == nil {
		return fmt.Errorf("qmm is not initialized")
	}

	fmt.Println("Adding mod zip", path)

	// copy zip to cache directory, and also look for qmm.zip

	r, err := zip.OpenReader(path)
	if err != nil {
		return fmt.Errorf("open reader: %w", err)
	}
	defer r.Close()

	yaml, err := findYamlInZip(r)
	if err != nil {
		return fmt.Errorf("find yaml in zip: %w", err)
	}

	fi, err := os.Stat("cache")
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Printf("Creating cache directory\n")

			err = os.MkdirAll("cache", os.ModePerm)
			if err != nil {
				return fmt.Errorf("mkdir cache: %w", err)
			}
		}
	} else if !fi.IsDir() {
		return fmt.Errorf("cache is not a directory")
	}

	outFile, err := os.OpenFile(fmt.Sprintf("cache/%s-%s.zip", yaml.ID, yaml.Version), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("open file: %w", err)
	}
	defer outFile.Close()

	zw := zip.NewWriter(outFile)
	defer zw.Close()

	for _, f := range r.File {
		rc, err := f.Open()
		if err != nil {
			return fmt.Errorf("file open: %w", err)
		}
		defer rc.Close()

		w, err := zw.Create(f.Name)
		if err != nil {
			return fmt.Errorf("create: %w", err)
		}

		var buf bytes.Buffer
		mw := io.MultiWriter(w, &buf)
		_, err = io.Copy(mw, rc)
		if err != nil {
			return fmt.Errorf("copy: %w", err)
		}
		if f.Name == yaml.Image {
			yaml.ImageData = buf.Bytes()
			yaml.imageBitmap = ico.Generate(yaml.ID, yaml.ImageData)
		}
	}

	for _, e := range q.Entries {
		if e.ID == yaml.ID {
			return fmt.Errorf("mod already already as %s, remove it", e.Name)
		}
	}

	yaml.IsEnabled = true
	q.Entries = append(q.Entries, yaml)
	err = rebuildModlist()
	if err != nil {
		return fmt.Errorf("rebuild modlist: %w", err)
	}

	err = q.saveConfig()
	if err != nil {
		return fmt.Errorf("save config: %w", err)
	}
	return nil
}

func rebuildModlist() error {
	q := instance
	if q == nil {
		return fmt.Errorf("qmm is not initialized")
	}

	fmt.Printf("Rebuilding modlist with %d entries\n", len(q.Entries))

	entries := make([]*component.ModViewEntry, len(instance.Entries))
	for i, e := range instance.Entries {
		fmt.Printf("%s: %t\n", e.Name, e.IsEnabled)
		mventry := &component.ModViewEntry{
			IsEnabled: e.IsEnabled,
			ID:        e.ID,
			Name:      e.Name,
			URL:       e.URL,
			Version:   e.Version,
		}
		if e.imageBitmap != nil {
			mventry.Icon = e.imageBitmap
		} else {
			mventry.Icon = ico.Grab("mod")
		}

		entries[i] = mventry
	}

	mw.SetModEntries(entries)
	return nil
}

func generateMod() error {

	if instance == nil {
		return fmt.Errorf("qmm is not initialized")
	}

	err := os.MkdirAll("cache", os.ModePerm)
	if err != nil {
		return fmt.Errorf("mkdir cache: %w", err)
	}

	numEntries := 0

	for _, e := range instance.Entries {
		if !e.IsEnabled {
			continue
		}
		path := fmt.Sprintf("cache/%s-%s.zip", e.ID, e.Version)
		r, err := zip.OpenReader(path)
		if err != nil {
			return fmt.Errorf("open reader: %w", err)
		}
		defer r.Close()

		_, err = findYamlInZip(r)
		if err != nil {
			return fmt.Errorf("find yaml in zip: %w", err)
		}

		for _, f := range r.File {
			fpath := filepath.Join("cache", fmt.Sprintf("%d", numEntries), f.Name)
			if f.FileInfo().IsDir() {
				os.MkdirAll(fpath, os.ModePerm)
				continue
			}

			if err := os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
				return fmt.Errorf("mkdir all: %w", err)
			}

			outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
			if err != nil {
				return fmt.Errorf("open file: %w", err)
			}

			rc, err := f.Open()
			if err != nil {
				return fmt.Errorf("file open: %w", err)
			}

			_, err = io.Copy(outFile, rc)
			if err != nil {
				return fmt.Errorf("copy: %w", err)
			}

			outFile.Close()
			rc.Close()
		}

		numEntries++
	}

	fmt.Println("Generated", numEntries, "entries")
	q := quail.New()
	q.Wld = wce.New("grass.wld")

	if q.Assets == nil {
		q.Assets = make(map[string][]byte)
	}

	for i := range numEntries {

		path := filepath.Join("cache", fmt.Sprintf("%d", i))

		q2 := quail.New()
		err := q2.DirRead(path)
		if err != nil {
			return fmt.Errorf("dir read: %w", err)
		}

		for _, d := range q2.Wld.ActorDefs {
			q.Wld.ActorDefs = append(q.Wld.ActorDefs, d)
		}
		for _, d := range q2.Wld.ActorInsts {
			q.Wld.ActorInsts = append(q.Wld.ActorInsts, d)
		}
		for _, d := range q2.Wld.AmbientLights {
			q.Wld.AmbientLights = append(q.Wld.AmbientLights, d)
		}
		for _, d := range q2.Wld.BlitSpriteDefs {
			q.Wld.BlitSpriteDefs = append(q.Wld.BlitSpriteDefs, d)
		}
		for _, d := range q2.Wld.DMSpriteDef2s {
			q.Wld.DMSpriteDef2s = append(q.Wld.DMSpriteDef2s, d)
		}
		for _, d := range q2.Wld.DMSpriteDefs {
			q.Wld.DMSpriteDefs = append(q.Wld.DMSpriteDefs, d)
		}
		for _, d := range q2.Wld.DMTrackDef2s {
			q.Wld.DMTrackDef2s = append(q.Wld.DMTrackDef2s, d)
		}
		for _, d := range q2.Wld.HierarchicalSpriteDefs {
			q.Wld.HierarchicalSpriteDefs = append(q.Wld.HierarchicalSpriteDefs, d)
		}
		for _, d := range q2.Wld.LightDefs {
			q.Wld.LightDefs = append(q.Wld.LightDefs, d)
		}
		for _, d := range q2.Wld.MaterialDefs {
			q.Wld.MaterialDefs = append(q.Wld.MaterialDefs, d)
		}
		for _, d := range q2.Wld.MaterialPalettes {
			q.Wld.MaterialPalettes = append(q.Wld.MaterialPalettes, d)
		}
		for _, d := range q2.Wld.ParticleCloudDefs {
			q.Wld.ParticleCloudDefs = append(q.Wld.ParticleCloudDefs, d)
		}
		for _, d := range q2.Wld.PointLights {
			q.Wld.PointLights = append(q.Wld.PointLights, d)
		}
		for _, d := range q2.Wld.PolyhedronDefs {
			q.Wld.PolyhedronDefs = append(q.Wld.PolyhedronDefs, d)
		}
		for _, d := range q2.Wld.Regions {
			q.Wld.Regions = append(q.Wld.Regions, d)
		}
		for _, d := range q2.Wld.RGBTrackDefs {
			q.Wld.RGBTrackDefs = append(q.Wld.RGBTrackDefs, d)
		}
		for _, d := range q2.Wld.SimpleSpriteDefs {
			q.Wld.SimpleSpriteDefs = append(q.Wld.SimpleSpriteDefs, d)
		}
		for _, d := range q2.Wld.Sprite2DDefs {
			q.Wld.Sprite2DDefs = append(q.Wld.Sprite2DDefs, d)
		}
		for _, d := range q2.Wld.Sprite3DDefs {
			q.Wld.Sprite3DDefs = append(q.Wld.Sprite3DDefs, d)
		}
		for _, d := range q2.Wld.TrackDefs {
			q.Wld.TrackDefs = append(q.Wld.TrackDefs, d)
		}
		for _, d := range q2.Wld.TrackInstances {
			q.Wld.TrackInstances = append(q.Wld.TrackInstances, d)
		}
		// for k, v := range q2.Wld.VariationMaterialDefs {
		// 	q.Wld.VariationMaterialDefs[k] = append(q.Wld.VariationMaterialDefs[k], v...)
		// }
		for _, d := range q2.Wld.WorldTrees {
			q.Wld.WorldTrees = append(q.Wld.WorldTrees, d)
		}
		for _, d := range q2.Wld.Zones {
			q.Wld.Zones = append(q.Wld.Zones, d)
		}
		for _, d := range q2.Wld.AniDefs {
			q.Wld.AniDefs = append(q.Wld.AniDefs, d)
		}
		for _, d := range q2.Wld.MdsDefs {
			q.Wld.MdsDefs = append(q.Wld.MdsDefs, d)
		}
		for _, d := range q2.Wld.ModDefs {
			q.Wld.ModDefs = append(q.Wld.ModDefs, d)
		}
		for _, d := range q2.Wld.TerDefs {
			q.Wld.TerDefs = append(q.Wld.TerDefs, d)
		}
		for _, d := range q2.Wld.LayDefs {
			q.Wld.LayDefs = append(q.Wld.LayDefs, d)
		}
		for _, d := range q2.Wld.PtsDefs {
			q.Wld.PtsDefs = append(q.Wld.PtsDefs, d)
		}
		for _, d := range q2.Wld.PrtDefs {
			q.Wld.PrtDefs = append(q.Wld.PrtDefs, d)
		}

		for k, v := range q2.Assets {
			q.Assets[k] = v
		}

		q2.Close()
	}

	grassDir, err := grass.Assets.ReadDir("assets")
	if err != nil {
		return fmt.Errorf("read dir: %w", err)
	}

	for _, file := range grassDir {
		q.Assets[file.Name()], err = grass.Assets.ReadFile("assets/" + file.Name())
		if err != nil {
			return fmt.Errorf("read %s: %w", file.Name(), err)
		}

	}

	err = q.PfsWrite(0, 0, "grass.s3d")
	if err != nil {
		return fmt.Errorf("pfs write: %w", err)
	}
	return nil
}

func findYamlInZip(r *zip.ReadCloser) (*QmmYamlEntry, error) {
	for _, f := range r.File {
		if f.Name != "qmm.yaml" {
			continue
		}
		rc, err := f.Open()
		if err != nil {
			return nil, fmt.Errorf("file open: %w", err)
		}
		defer rc.Close()

		decoder := yaml.NewDecoder(rc)
		yaml := &QmmYaml{}
		err = decoder.Decode(yaml)
		if err != nil {
			return nil, fmt.Errorf("decode: %w", err)
		}
		if len(yaml.Entry) < 1 {
			return nil, fmt.Errorf("no entries found in qmm.yaml")
		}

		entry := &yaml.Entry[0]
		if entry.ID == "" {
			return nil, fmt.Errorf("id not found in qmm.yaml")
		}
		if entry.Name == "" {
			return nil, fmt.Errorf("name not found in qmm.yaml")
		}

		return entry, nil
	}
	return nil, fmt.Errorf("qmm.yaml not found")
}

func onRemoveMod(modID string) {
	q := instance
	if q == nil {
		dialog.ShowMessageBox("Error", "qmm is not initialized", true)
		return
	}

	for i, e := range q.Entries {
		if e.ID != modID {
			continue
		}
		q.Entries = append(q.Entries[:i], q.Entries[i+1:]...)
		break
	}
	err := rebuildModlist()
	if err != nil {
		dialog.ShowMessageBox("Error", err.Error(), true)
		return
	}
	err = q.saveConfig()
	if err != nil {
		dialog.ShowMessageBox("Error", err.Error(), true)
		return
	}

}

func onGenerateMod() {
	err := generateMod()
	if err != nil {
		dialog.ShowMessageBox("Error", err.Error(), true)
		return
	}
}

// loadConfig loads from quail-mod-manager.bin via gob
func (q *Qmm) loadConfig() error {

	r, err := os.Open("quail-mod-manager.bin")
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("open: %w", err)
	}
	defer r.Close()

	nq := &Qmm{}
	err = gob.NewDecoder(r).Decode(nq)
	if err != nil {
		return fmt.Errorf("decode gob: %w", err)
	}
	q.Entries = nq.Entries
	return nil
}

// saveConfig saves to quail-mod-manager.bin via gob
func (q *Qmm) saveConfig() error {
	w, err := os.OpenFile("quail-mod-manager.bin", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("open file: %w", err)
	}
	defer w.Close()

	err = gob.NewEncoder(w).Encode(q)
	if err != nil {
		return fmt.Errorf("encode: %w", err)
	}
	return nil
}

func onModURL() {
	var diaWlk *walk.Dialog
	var acceptBtn *walk.PushButton
	var cancelBtn *walk.PushButton
	var urlLineEdit *walk.LineEdit

	dia := cpl.Dialog{
		AssignTo:      &diaWlk,
		Title:         "Import Mod URL",
		DefaultButton: &acceptBtn,
		CancelButton:  &cancelBtn,
		MinSize:       cpl.Size{Width: 300, Height: 200},
		Layout:        cpl.VBox{},
		Children: []cpl.Widget{
			cpl.Composite{
				Layout: cpl.Grid{Columns: 2},
				Children: []cpl.Widget{
					&cpl.Label{
						Text: "https://github.com/",
					},
					&cpl.LineEdit{
						AssignTo: &urlLineEdit,
						Text:     cpl.Bind("url"),
						OnKeyPress: func(key walk.Key) {
							if key != walk.KeyReturn {
								return
							}
							url := urlLineEdit.Text()
							fmt.Println("URL", url)
							if url == "" {
								dialog.ShowMessageBox("Error", "URL is empty", true)
								return
							}

							err := AddModURL(url)
							if err != nil {
								dialog.ShowMessageBox("Error", err.Error(), true)
								return
							}
							diaWlk.Accept()
						},
					},
				},
			},
			&cpl.Composite{
				Layout: cpl.HBox{},
				Children: []cpl.Widget{
					&cpl.PushButton{
						AssignTo: &acceptBtn,
						Text:     "Accept",
						OnClicked: func() {
							url := urlLineEdit.Text()
							fmt.Println("URL", url)
							if url == "" {
								dialog.ShowMessageBox("Error", "URL is empty", true)
								return
							}

							err := AddModURL(url)
							if err != nil {
								dialog.ShowMessageBox("Error", err.Error(), true)
								return
							}
							diaWlk.Accept()
						},
					},
					&cpl.PushButton{
						AssignTo:  &cancelBtn,
						Text:      "Cancel",
						OnClicked: func() { diaWlk.Cancel() },
					},
				},
			},
		},
	}

	err := dia.Create(mw.Instance().MainWindowWlk)
	if err != nil {
		dialog.ShowMessageBox("Error", err.Error(), true)
		return
	}
	errCode, err := dia.Run(mw.Instance().MainWindowWlk)
	if err != nil {
		dialog.ShowMessageBox("Error", err.Error(), true)
		return
	}

	if errCode != walk.DlgCmdOK {
		return
	}

}

func AddModURL(url string) error {
	q := instance
	if q == nil {
		return fmt.Errorf("qmm is not initialized")
	}

	url = "https://github.com/" + url
	fmt.Println("Adding mod url", url)

	if !strings.HasPrefix(url, "https://github.com/") {
		return fmt.Errorf("unsupported URL, only GitHub URLs are supported")
	}

	parts := strings.Split(url, "/")
	if len(parts) < 5 {
		return fmt.Errorf("invalid GitHub URL")
	}

	owner := parts[3]
	repo := parts[4]

	apiURL := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", owner, repo)
	resp, err := http.Get(apiURL)
	if err != nil {
		return fmt.Errorf("failed to fetch latest release: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to fetch latest release: %s", resp.Status)
	}

	var release struct {
		Assets []struct {
			BrowserDownloadURL string `json:"browser_download_url"`
			Name               string `json:"name"`
		} `json:"assets"`
	}

	err = json.NewDecoder(resp.Body).Decode(&release)
	if err != nil {
		return fmt.Errorf("failed to decode release JSON: %w", err)
	}

	var zipURL string
	for _, asset := range release.Assets {
		if strings.HasSuffix(asset.Name, ".zip") {
			zipURL = asset.BrowserDownloadURL
			break
		}
	}

	if zipURL == "" {
		return fmt.Errorf("no zip asset found in the latest release")
	}

	resp, err = http.Get(zipURL)
	if err != nil {
		return fmt.Errorf("failed to download zip: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download zip: %s", resp.Status)
	}

	tempFile, err := os.CreateTemp("", "mod-*.zip")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	defer tempFile.Close()

	_, err = io.Copy(tempFile, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to save zip: %w", err)
	}

	err = AddModZip(tempFile.Name())
	if err != nil {
		return fmt.Errorf("failed to add mod zip: %w", err)
	}

	return nil
}

func onEnableMod(modID string, state bool) {
	q := instance
	if q == nil {
		dialog.ShowMessageBox("Error", "qmm is not initialized", true)
		return
	}

	for _, e := range q.Entries {
		if e.ID != modID {
			continue
		}
		e.IsEnabled = state
		break
	}
	err := rebuildModlist()
	if err != nil {
		dialog.ShowMessageBox("Error", err.Error(), true)
		return
	}
	err = q.saveConfig()
	if err != nil {
		dialog.ShowMessageBox("Error", err.Error(), true)
		return
	}
	fmt.Printf("Saved change")
}
