package qmm

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"quail-mod-manager/dialog"
	"quail-mod-manager/grass"
	"quail-mod-manager/handler"

	"github.com/xackery/quail/quail"
	"github.com/xackery/quail/wce"
)

func New() error {
	handler.ImportModZipSubscribe(onModZip)
	return nil
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
	dialog.ShowMessageBox("Success", "Selected: "+path, false)
}

func generateMod() error {

	cacheFiles, err := os.ReadDir("../bin/mods/")
	if err != nil {
		return fmt.Errorf("read dir: %w", err)
	}
	cacheOut := "cache/"
	err = os.RemoveAll(cacheOut)
	if err != nil {
		return fmt.Errorf("remove all: %w", err)
	}
	numEntries := 0

	for _, file := range cacheFiles {
		if file.IsDir() {
			continue
		}
		ext := filepath.Ext(file.Name())
		if ext != ".zip" {
			continue
		}

		r, err := zip.OpenReader("../bin/mods/" + file.Name())
		if err != nil {
			return fmt.Errorf("open reader: %w", err)
		}
		defer r.Close()

		hasMetafile := false
		for _, f := range r.File {
			if f.Name != "qmm.yaml" {
				continue
			}
			hasMetafile = true
			break
		}
		if !hasMetafile {
			continue
		}

		for _, f := range r.File {
			fpath := filepath.Join(cacheOut, fmt.Sprintf("%d", numEntries), f.Name)
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

		path := filepath.Join(cacheOut, fmt.Sprintf("%d", i))

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
