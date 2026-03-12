package gdrive

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
)

// DriveFileInfo holds basic metadata about a file in Google Drive.
type DriveFileInfo struct {
	ID       string
	Name     string
	MimeType string
}

// DriveService abstracts Google Drive file operations for easy mocking in tests.
type DriveService interface {
	UploadFile(ctx context.Context, filename string, content io.Reader, mimeType string, folderID string) (fileID string, err error)
	FindOrCreateFolder(ctx context.Context, name string, parentID string) (folderID string, err error)
	GetFile(ctx context.Context, fileID string) (content io.ReadCloser, mimeType string, filename string, err error)
	ListFilesInFolder(ctx context.Context, folderID string) ([]DriveFileInfo, error)
}

type driveService struct {
	service *drive.Service
}

// NewDriveService creates a DriveService using the provided authenticated HTTP client.
func NewDriveService(ctx context.Context, httpClient *http.Client) (DriveService, error) {
	srv, err := drive.NewService(ctx, option.WithHTTPClient(httpClient))
	if err != nil {
		return nil, fmt.Errorf("unable to create Drive service: %w", err)
	}
	return &driveService{service: srv}, nil
}

// UploadFile uploads a file to Google Drive and returns its file ID.
// If folderID is non-empty, the file is placed in that folder.
func (d *driveService) UploadFile(ctx context.Context, filename string, content io.Reader, mimeType string, folderID string) (string, error) {
	file := &drive.File{
		Name:     filename,
		MimeType: mimeType,
	}
	if folderID != "" {
		file.Parents = []string{folderID}
	}

	created, err := d.service.Files.Create(file).Media(content).SupportsAllDrives(true).Context(ctx).Do()
	if err != nil {
		return "", fmt.Errorf("unable to upload file to Drive: %w", err)
	}
	return created.Id, nil
}

// GetFile downloads a file from Google Drive by its file ID.
// Returns the file content stream, MIME type, and original filename.
func (d *driveService) GetFile(ctx context.Context, fileID string) (io.ReadCloser, string, string, error) {
	// Get file metadata first
	meta, err := d.service.Files.Get(fileID).Fields("name, mimeType").SupportsAllDrives(true).Context(ctx).Do()
	if err != nil {
		return nil, "", "", fmt.Errorf("unable to get file metadata from Drive: %w", err)
	}

	// Download file content
	resp, err := d.service.Files.Get(fileID).SupportsAllDrives(true).Context(ctx).Download()
	if err != nil {
		return nil, "", "", fmt.Errorf("unable to download file from Drive: %w", err)
	}

	return resp.Body, meta.MimeType, meta.Name, nil
}

// FindOrCreateFolder looks for an existing folder with the given name inside parentID.
// If not found, it creates one. Returns the folder ID.
func (d *driveService) FindOrCreateFolder(ctx context.Context, name string, parentID string) (string, error) {
	// Search for existing folder
	q := fmt.Sprintf("name = '%s' and mimeType = 'application/vnd.google-apps.folder' and trashed = false", name)
	if parentID != "" {
		q += fmt.Sprintf(" and '%s' in parents", parentID)
	}

	result, err := d.service.Files.List().Q(q).Fields("files(id)").SupportsAllDrives(true).IncludeItemsFromAllDrives(true).Context(ctx).Do()
	if err != nil {
		return "", fmt.Errorf("unable to search for folder: %w", err)
	}

	if len(result.Files) > 0 {
		return result.Files[0].Id, nil
	}

	// Create the folder
	folder := &drive.File{
		Name:     name,
		MimeType: "application/vnd.google-apps.folder",
	}
	if parentID != "" {
		folder.Parents = []string{parentID}
	}

	created, err := d.service.Files.Create(folder).SupportsAllDrives(true).Context(ctx).Do()
	if err != nil {
		return "", fmt.Errorf("unable to create folder: %w", err)
	}
	return created.Id, nil
}

// ListFilesInFolder lists all non-trashed files in the given folder.
// It handles pagination to return all files, not just the first page.
func (d *driveService) ListFilesInFolder(ctx context.Context, folderID string) ([]DriveFileInfo, error) {
	q := fmt.Sprintf("'%s' in parents and trashed = false", folderID)
	var files []DriveFileInfo
	pageToken := ""
	for {
		call := d.service.Files.List().Q(q).Fields("nextPageToken, files(id, name, mimeType)").SupportsAllDrives(true).IncludeItemsFromAllDrives(true).Context(ctx)
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		result, err := call.Do()
		if err != nil {
			return nil, fmt.Errorf("unable to list files in folder: %w", err)
		}
		for _, f := range result.Files {
			files = append(files, DriveFileInfo{
				ID:       f.Id,
				Name:     f.Name,
				MimeType: f.MimeType,
			})
		}
		if result.NextPageToken == "" {
			break
		}
		pageToken = result.NextPageToken
	}
	return files, nil
}
