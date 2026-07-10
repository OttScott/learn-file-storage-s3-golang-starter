package main

import (
	"fmt"
	"io"
	"os"
	"bytes"
	"net/http"

	"github.com/bootdotdev/learn-file-storage-s3-golang-starter/internal/auth"
	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerUploadThumbnail(writer http.ResponseWriter, req *http.Request) {
	videoIDString := req.PathValue("videoID")
	videoID, err := uuid.Parse(videoIDString)
	if err != nil {
		respondWithError(writer, http.StatusBadRequest, "Invalid ID", err)
		return
	}

	token, err := auth.GetBearerToken(req.Header)
	if err != nil {
		respondWithError(writer, http.StatusUnauthorized, "Couldn't find JWT", err)
		return
	}

	userID, err := auth.ValidateJWT(token, cfg.jwtSecret)
	if err != nil {
		respondWithError(writer, http.StatusUnauthorized, "Couldn't validate JWT", err)
		return
	}


	fmt.Println("uploading thumbnail for video", videoID, "by user", userID)

	// TODO: implement the upload here
	const maxMemory = 10 << 20 // 10 MB
	req.ParseMultipartForm(maxMemory)

	file, header, err := req.FormFile("thumbnail")
	if err != nil {
		respondWithError(writer, http.StatusBadRequest, "Unable to parse form file", err)
		return
	}
	defer file.Close()

	mediaType := header.Header.Get("Content-Type")

	// Read the file data into a byte slice using io.ReadAll
	fileData, err := io.ReadAll(file)
	if err != nil {
		respondWithError(writer, http.StatusInternalServerError, "Unable to read file data", err)
		return
	}

	videoInfo, err := cfg.db.GetVideo(videoID)
	if err != nil {
		respondWithError(writer, http.StatusInternalServerError, "Unable to get video", err)
		return
	}
	if videoInfo.UserID != userID {
		respondWithError(writer, http.StatusUnauthorized, "You do not have permission to upload a thumbnail for this video", nil)
		return
	}

	fileExtension := ""
	switch mediaType {
	case "image/jpeg":
		fileExtension = "jpg"
	case "image/png":
		fileExtension = "png"
	case "image/gif":
		fileExtension = "gif"
	default:
		respondWithError(writer, http.StatusBadRequest, "Unsupported media type", nil)
		return
	}
	filePath := fmt.Sprintf("%s/%s.%s", cfg.assetsRoot, videoID, fileExtension)

	outFile, err := os.Create(filePath)
	if err != nil {
		respondWithError(writer, http.StatusInternalServerError, "Unable to create file", err)
		return
	}
	defer outFile.Close()

	
	_, err = io.Copy(outFile, bytes.NewReader(fileData))
	if err != nil {
		respondWithError(writer, http.StatusInternalServerError, "Unable to write file data", err)
		return
	}

	thumbnailURL := fmt.Sprintf("http://%s:%s/%s/%s.%s", "localhost", "8091", cfg.assetsRoot, videoID, fileExtension)
	videoInfo.ThumbnailURL = &thumbnailURL
	err = cfg.db.UpdateVideo(videoInfo)
	if err != nil {
		respondWithError(writer, http.StatusInternalServerError, "Unable to update video with thumbnail URL", err)
		return
	}

	// Respond with updated JSON of the video's metadata. Use the provided respondWithJSON function and pass it the updated database.Video struct to marshal.
	respondWithJSON(writer, http.StatusOK, videoInfo)
}
