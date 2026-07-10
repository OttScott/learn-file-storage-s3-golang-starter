package main

import (
	"encoding/base64"
	"fmt"
	"io"
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

	dataURL := fmt.Sprintf("data:%s;base64,%s", mediaType, base64.StdEncoding.EncodeToString(fileData))

	videoInfo.ThumbnailURL = &dataURL
	err = cfg.db.UpdateVideo(videoInfo)
	if err != nil {
		respondWithError(writer, http.StatusInternalServerError, "Unable to update video with thumbnail URL", err)
		return
	}

	// Respond with updated JSON of the video's metadata. Use the provided respondWithJSON function and pass it the updated database.Video struct to marshal.
	respondWithJSON(writer, http.StatusOK, videoInfo)
}
