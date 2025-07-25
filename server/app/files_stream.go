package app

import (
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/mattermost/focalboard/server/model"
	"github.com/mattermost/focalboard/server/utils"
	"github.com/mattermost/mattermost/server/public/shared/mlog"
)

// SaveFileStreaming saves a file using streaming to avoid loading entire file into memory
// This is a memory-efficient version of SaveFile for large files
func (a *App) SaveFileStreaming(reader io.Reader, teamID, boardID, filename string, asTemplate bool) (string, error) {
	// NOTE: File extension includes the dot
	fileExtension := strings.ToLower(filepath.Ext(filename))
	if fileExtension == ".jpeg" {
		fileExtension = ".jpg"
	}

	createdFilename := utils.NewID(utils.IDTypeNone)
	newFileName := fmt.Sprintf(`%s%s`, createdFilename, fileExtension)
	if asTemplate {
		newFileName = filename
	}
	filePath := getDestinationFilePath(asTemplate, teamID, boardID, newFileName)

	// Use a limited reader to prevent memory exhaustion
	// Process in 1MB chunks to stay well under memory limits
	limitedReader := &io.LimitedReader{
		R: reader,
		N: 100 * 1024 * 1024, // 100MB max per file
	}

	// Create a buffered reader with smaller buffer size to control memory usage
	bufSize := 64 * 1024 // 64KB chunks
	bufferedReader := &BufferedCopyReader{
		reader: limitedReader,
		bufSize: bufSize,
	}

	fileSize, appErr := a.filesBackend.WriteFile(bufferedReader, filePath)
	if appErr != nil {
		return "", fmt.Errorf("unable to store the file in the files storage: %w", appErr)
	}

	// Check if we hit the limit
	if limitedReader.N == 0 {
		a.logger.Warn("File size limit reached during upload",
			mlog.String("filename", filename),
			mlog.Int("maxSize", 100*1024*1024),
		)
	}

	fileInfo := model.NewFileInfo(filename)
	fileInfo.Id = getFileInfoID(createdFilename)
	fileInfo.Path = filePath
	fileInfo.Size = fileSize

	err := a.store.SaveFileInfo(fileInfo)
	if err != nil {
		return "", err
	}

	return newFileName, nil
}

// BufferedCopyReader wraps an io.Reader to control memory usage during copying
type BufferedCopyReader struct {
	reader  io.Reader
	bufSize int
}

func (b *BufferedCopyReader) Read(p []byte) (n int, err error) {
	// Limit read size to our buffer size to control memory usage
	if len(p) > b.bufSize {
		p = p[:b.bufSize]
	}
	return b.reader.Read(p)
}