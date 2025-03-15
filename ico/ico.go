package ico

import (
	"bytes"
	"fmt"
	"image"
	"image/png"
	"path/filepath"
	"strings"

	_ "embed"

	ico "github.com/biessek/golang-ico"
	"github.com/malashin/dds"
	"github.com/sergeymakinen/go-bmp"
	"github.com/xackery/wlk/walk"
	"golang.org/x/image/draw"
)

var (
	//go:embed assets/unk.ico
	unkIco []byte
	//go:embed assets/ani.ico
	aniIco []byte
	//go:embed assets/mds.ico
	mdsIco []byte
	//go:embed assets/lay.ico
	layIco []byte
	//go:embed assets/mod.ico
	modIco []byte
	//go:embed assets/pts.ico
	ptsIco []byte
	//go:embed assets/prt.ico
	prtIco []byte
	//go:embed assets/zon.ico
	zonIco []byte
	//go:embed assets/ter.ico
	terIco []byte
	//go:embed assets/lit.ico
	litIco []byte
	//go:embed assets/wld.ico
	wldIco []byte
	//go:embed assets/bon.ico
	bonIco []byte
	//go:embed assets/mat.ico
	matIco []byte
	//go:embed assets/tri.ico
	triIco []byte
	//go:embed assets/ver.ico
	verIco []byte
	//go:embed assets/header.ico
	headerIco []byte
	//go:embed assets/obj.ico
	objIco []byte
	//go:embed assets/region.ico
	regionIco []byte
	//go:embed assets/preview.ico
	previewIco []byte
	icos       map[string]*walk.Bitmap

	icoMap = map[string][]byte{
		".ani":    aniIco,
		".mds":    mdsIco,
		".lay":    layIco,
		".mod":    modIco,
		".unk":    unkIco,
		".pts":    ptsIco,
		".prt":    prtIco,
		".zon":    zonIco,
		".ter":    terIco,
		".lit":    litIco,
		".wld":    wldIco,
		".bon":    bonIco,
		".mat":    matIco,
		".tri":    triIco,
		".ver":    verIco,
		".obj":    objIco,
		"header":  headerIco,
		"region":  regionIco,
		"preview": previewIco,
	}
)

func init() {

	icos = make(map[string]*walk.Bitmap)
	for ext, icoData := range icoMap {
		img, err := ico.Decode(bytes.NewReader(icoData))
		if err != nil {
			fmt.Printf("Failed to decode %s: %s\n", ext, err.Error())
			continue
		}
		bmp, err := walk.NewBitmapFromImageForDPI(img, 96)
		if err != nil {
			fmt.Printf("Failed to create bitmap from image: %s\n", err.Error())
			continue
		}
		icos[ext] = bmp
	}
}

// Grab returns a walk.Bitmap for a given icon
func Grab(name string) *walk.Bitmap {
	bmp, ok := icos[name]
	if !ok {
		return icos[".unk"]
	}
	return bmp
}

func Generate(name string, data []byte) *walk.Bitmap {
	var err error

	wBmp, ok := icos[name]
	if ok {
		return wBmp
	}

	ext := strings.ToLower(filepath.Ext(name))
	defer func() {
		if err != nil {
			fmt.Printf("GenerateIcon: %s", err)
			return
		}
	}()

	wBmp = Grab(ext)
	if wBmp != nil && ext != ".unk" {
		return wBmp
	}

	unkImg := Grab(".unk")

	var img image.Image
	if ext == ".dds" {
		img, err = dds.Decode(bytes.NewReader(data))
		if err != nil {
			err = fmt.Errorf("dds.Decode %s: %w", name, err)
			return unkImg
		}
		dst := image.NewRGBA(image.Rect(0, 0, img.Bounds().Max.X/2, img.Bounds().Max.Y/2))
		draw.NearestNeighbor.Scale(dst, image.Rect(0, 0, 16, 16), img, img.Bounds(), draw.Over, nil)

		wBmp, err = walk.NewBitmapFromImageForDPI(dst, 96)
		if err != nil {
			err = fmt.Errorf("new bitmap from image for dpi: %s", err)
			return unkImg
		}
		return wBmp
	}

	if ext == ".png" {
		img, err = png.Decode(bytes.NewReader(data))
		if err != nil {
			err = fmt.Errorf("png.Decode %s: %w", name, err)
			return unkImg
		}
		dst := image.NewRGBA(image.Rect(0, 0, img.Bounds().Max.X/2, img.Bounds().Max.Y/2))
		draw.NearestNeighbor.Scale(dst, image.Rect(0, 0, 16, 16), img, img.Bounds(), draw.Over, nil)

		wBmp, err = walk.NewBitmapFromImageForDPI(dst, 96)
		if err != nil {
			err = fmt.Errorf("new bitmap from image for dpi: %s", err)
			return unkImg
		}
		return wBmp
	}
	if ext == ".bmp" {
		buf := new(bytes.Buffer)
		_, err = buf.ReadFrom(bytes.NewReader(data))
		if err != nil {
			err = fmt.Errorf("buf read from: %w", err)
			return unkImg
		}
		var img image.Image
		if string(buf.Bytes()[0:3]) == "DDS" {
			img, err = dds.Decode(bytes.NewReader(data))
			if err != nil {
				err = fmt.Errorf("dds.Decode %s: %w", name, err)
				return unkImg
			}
		} else {
			img, err = bmp.Decode(bytes.NewReader(data))
			if err != nil {
				err = fmt.Errorf("bmp.Decode %s: %w", name, err)
				return unkImg
			}
		}
		dst := image.NewRGBA(image.Rect(0, 0, img.Bounds().Max.X/2, img.Bounds().Max.Y/2))
		draw.NearestNeighbor.Scale(dst, image.Rect(0, 0, 16, 16), img, img.Bounds(), draw.Over, nil)

		wBmp, err = walk.NewBitmapFromImageForDPI(dst, 96)
		if err != nil {
			err = fmt.Errorf("new bitmap from image for dpi: %w", err)
			return unkImg
		}
		return wBmp
	}

	fmt.Println("unk ext", ext, unkImg)

	return unkImg
}

// Clear is used to flush an ico or generate cache
func Clear(name string) {
	_, ok := icoMap[name]
	if ok {
		return
	}
	delete(icos, name)
}
