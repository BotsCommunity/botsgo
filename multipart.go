package botsgo

import (
	"bytes"
	"io"
	"mime/multipart"
	"os"
)

type MultiPart struct {
	r      *Requester
	buffer *bytes.Buffer
	writer *multipart.Writer
}

func (r *Requester) NewMultiPart() *MultiPart {
	buffer := new(bytes.Buffer)
	writer := multipart.NewWriter(buffer)

	return &MultiPart{
		r:      r,
		buffer: buffer,
		writer: writer,
	}
}

func (mp *MultiPart) SetFormField(fieldName string, value []byte) error {
	writer, err := mp.writer.CreateFormField(fieldName)
	if err != nil {
		return err
	}

	if _, err := writer.Write(value); err != nil {
		return err
	}

	return nil
}

func (mp *MultiPart) SetFormFile(fieldName string, fileName string, file *os.File) error {
	writer, err := mp.writer.CreateFormFile(fieldName, fileName)
	if err != nil {
		return err
	}

	if _, err := io.Copy(writer, file); err != nil {
		return err
	}

	return nil
}

func (mp *MultiPart) Buffer() (*bytes.Buffer, string, error) {
	err := mp.writer.Close()
	return mp.buffer, mp.writer.FormDataContentType(), err
}
