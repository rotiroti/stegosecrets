package encrypt

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/enrichman/stegosecrets/pkg/file"
	"github.com/enrichman/stegosecrets/pkg/image"
	sss "github.com/enrichman/stegosecrets/pkg/stego"
)

type Encrypter struct {
	Parts     int
	Threshold int
}

func EncryptFile(filename string) {

}

func (e *Encrypter) Encrypt(reader io.Reader) error {
	masterKey, err := sss.GenerateMasterKey()
	if err != nil {
		return err
	}
	file.WriteKey(masterKey, "out/file.aes")

	message, err := io.ReadAll(reader)
	if err != nil {
		return err
	}
	encryptedMessage, err := sss.Encrypt(masterKey, message)
	if err != nil {
		return err
	}

	file.WriteFile(encryptedMessage, "out/file.aes")
	file.WriteChecksum(encryptedMessage, "out/file.aes")

	if e.Parts > 1 {
		parts, err := sss.Split(masterKey, e.Parts, e.Threshold)
		if err != nil {
			return err
		}
		return encodePartsInImages(parts)
	}

	return nil
}

func encodePartsInImages(parts []sss.Part) error {
	dir := "images"
	// get images
	files, err := os.ReadDir(dir)
	if err != nil {
		return err
	}
	path, err := filepath.Abs(dir)
	if err != nil {
		return err
	}

	fmt.Printf("found %d images\n", len(files))

	images := []fs.DirEntry{}
	for _, f := range files {
		switch filepath.Ext(f.Name()) {
		case ".jpg", ".jpeg", ".png":
			images = append(images, f)
			fmt.Printf(" -  %s\n", f.Name())
		}
	}

	// TODO if parts > len(images) add same image

	fmt.Println("Encrypted parts:")

	for i, part := range parts {
		fmt.Printf(" %d) %s\n", i, part.Base64())

		outName := fmt.Sprintf("out/%d", i)
		if len(images) > 0 {
			imagePath := filepath.Join(path, images[i].Name())
			outName = fmt.Sprintf("%s%s", outName, filepath.Ext(images[i].Name()))
			image.EncodeSecretFromFile(part.Bytes(), imagePath, outName)
			file.WriteFileChecksum(outName)
		}

		file.WriteKey(part.Bytes(), outName)
	}

	return nil
}
