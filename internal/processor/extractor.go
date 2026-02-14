package processor

import (
	"archive/zip"
	"encoding/xml"
	"fmt"
	"my-go-vibe/internal/domain"
	"io"
)

// ExtractContent reads a .docx file and converts it into a DocumentRequest.
func ExtractContent(docPath string) (*domain.DocumentRequest, error) {
	// Open the docx file as a zip archive
	r, err := zip.OpenReader(docPath)
	if err != nil {
		return nil, err
	}
	defer r.Close()

	// Find the document.xml file
	var docFile *zip.File
	for _, f := range r.File {
		if f.Name == "word/document.xml" {
			docFile = f
			break
		}
	}

	if docFile == nil {
		return nil, fmt.Errorf("document.xml not found in %s", docPath)
	}

	// Open the document.xml file
	rc, err := docFile.Open()
	if err != nil {
		return nil, err
	}
	defer rc.Close()

	// Decode the XML and extract the content
	return extractContentFromXML(rc)
}

func extractContentFromXML(reader io.Reader) (*domain.DocumentRequest, error) {
	decoder := xml.NewDecoder(reader)
	var req domain.DocumentRequest
	var paraCount int

	for {
		token, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		if start, ok := token.(xml.StartElement); ok && start.Name.Local == "p" {
			paraCount++
			text, err := extractTextFromParagraph(decoder)
			if err != nil {
				return nil, err
			}

			req.Content = append(req.Content, domain.Block{
				ID:   fmt.Sprintf("p_%d", paraCount),
				Type: "paragraph",
				Text: text,
			})
		}
	}

	return &req, nil
}

func extractTextFromParagraph(decoder *xml.Decoder) (string, error) {
	var text string
	for {
		token, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", err
		}

		switch t := token.(type) {
		case xml.CharData:
			text += string(t)
		case xml.EndElement:
			if t.Name.Local == "p" {
				return text, nil
			}
		}
	}
	return text, nil
}
