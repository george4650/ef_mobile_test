package usecase

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"myapp/internal/models"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/hirochachacha/go-smb2"
	ffmpeg_go "github.com/u2takey/ffmpeg-go"
)

type ScreenRecorderUseCases struct {
	smbSession *smb2.Session
	storeOra   ScreenRecorderOracle
	storePg    ScreenRecorderPostgres
	audioLink  string
}

func NewScreenRecorderCases(audioLink string, smbSession *smb2.Session, storePg ScreenRecorderPostgres, storeOra ScreenRecorderOracle) *ScreenRecorderUseCases {
	return &ScreenRecorderUseCases{
		smbSession: smbSession,
		storeOra:   storeOra,
		storePg:    storePg,
		audioLink:  audioLink,
	}
}

func (us *ScreenRecorderUseCases) DownloadVideo(ctx context.Context, videoId string) (string, string, error) {

	var (
		file                 *smb2.File
		videoFile            *os.File
		audioFile            *os.File
		videoFileName        string
		audioFileName        string
		localVideoFileName   string
		videoFilesDirectory  string = "tmp/videoFiles"
		audioFilesDirectory  string = "tmp/audioFiles"
		shareDirectory       string = "scrnrec"
		resultFilesDirectory string = "tmp/resultVideos" //Директория, содержащая все склеенные с аудио видео
	)

	//Найти видео по id
	video, err := us.storePg.GetVideo(ctx, videoId)
	if err != nil {
		switch {
		//Если видео нет
		case strings.Contains(err.Error(), "Video does not exist"):
			log.Error().Err(err)
			return "", "", errors.New("Видео не существует")
		default:
			log.Error().Err(err)
			return "", "", fmt.Errorf("Внутренняя ошибка сервера: %w", err)
		}
	}

	sessionId := strings.Replace(video.SessionId, ".", "_", -1)

	month := fmt.Sprint(int(video.CreatedAt.Month()))
	day := fmt.Sprint(video.CreatedAt.Day())
	if int(video.CreatedAt.Month()) < 10 {
		month = "0" + fmt.Sprint(int(video.CreatedAt.Month()))
	}
	if video.CreatedAt.Day() < 10 {
		day = "0" + fmt.Sprint(video.CreatedAt.Day())
	}
	//Формируем название по которому требуется найти видео
	if video.SessionId == "" || video.SessionId == "none" {
		videoFileName = fmt.Sprintf("%s/%d/%s/%s/%d-%d-%d.mkv", video.Login, video.CreatedAt.Year(), month, day, video.CreatedAt.Hour(), video.CreatedAt.Minute(), video.CreatedAt.Second())
	} else {
		videoFileName = fmt.Sprintf("%s/%d/%s/%s/%s.mkv", video.Login, video.CreatedAt.Year(), month, day, sessionId)
	}

	log.Print("videoFileName:", videoFileName)

	localVideoFileName = strings.Replace(videoFileName, "/", "_", -1)

	//Прежде чем приступить к поиску видео в samba, ищем его в кэше в папке tmp/resultVideos
	files, err := ioutil.ReadDir(resultFilesDirectory)
	if err != nil {
		log.Error().Err(err)
	}
	for _, file := range files {
		if localVideoFileName == file.Name() {
			log.Print("Видеофайл найден в кэше")
			return localVideoFileName, "", nil
		}
	}

	log.Print("Видеофайл в кэше не найден")

	if us.smbSession == nil {
		return "", "", errors.New("Не удалось подключиться по локальной сети")
	}

	//Переходим в share директорию
	fs, err := us.smbSession.Mount(shareDirectory)
	if err != nil {
		return "", "", err
	}
	defer fs.Umount()

	//Ищем видео в samba
	file, err = fs.Open(videoFileName)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return "", "", errors.New("Видеофайла не существует")
		}
		return "", "", fmt.Errorf("Внутренняя ошибка сервера: %w", err)
	}
	defer file.Close()

	log.Print("Видеофайл найден в samba")

	// Create a local file for writing the downloaded content
	videoFile, err = os.Create(fmt.Sprintf("%s/%s", videoFilesDirectory, localVideoFileName))
	if err != nil {
		return "", "", errors.New("Внутренняя ошибка сервера")
	}

	buffer := make([]byte, 8192)
	// Read and write the file content
	for {
		n, err := file.Read(buffer)
		if err != nil {
			break
		}
		videoFile.Write(buffer[:n])
	}
	videoFile.Close()

	fmt.Println("Video downloaded successfully!")

	//Если sessionId пустое то перемещаем скаченное видео в папку resultVideos и выходим из функции
	if video.SessionId == "" {
		err := os.Rename(fmt.Sprintf("./%s/%s", videoFilesDirectory, localVideoFileName), fmt.Sprintf("./%s/%s", resultFilesDirectory, localVideoFileName))
		if err != nil {
			return "", "", errors.New("Внутренняя ошибка сервера")
		}
		return localVideoFileName, "Аудио отсутствует", nil
	}

	//Иначе ищем аудиозапись и склеиваем с видео

	//Скачать аудио по ссылке http://d.audiorecord:4490@nextstorage.nextcontact.ru/{year}/{month}/{day}/{session_id}.ogg?noredirect=true
	url := fmt.Sprintf("%s/%d/%s/%s/%s.ogg?noredirect=true", us.audioLink, video.CreatedAt.Year(), month, day, video.SessionId)

	log.Print("URL:", url)

	info, err := http.Get(url)

	//Если запись не найдена
	if info.StatusCode == http.StatusNotFound {
		log.Print("Аудиозапись не найдена!")
		//перемещаем видео в папку resultVideos и выходим из функции
		err := os.Rename(fmt.Sprintf("./%s/%s", videoFilesDirectory, localVideoFileName), fmt.Sprintf("./%s/%s", resultFilesDirectory, localVideoFileName))
		if err != nil {
			return "", "", fmt.Errorf("Внутренняя ошибка сервера: %w", err)
		}
		return localVideoFileName, "Не удалось найти аудиофайл в хранилище", nil
	}
	if err != nil {
		log.Error().Err(err)
		//перемещаем видео в папку resultVideos и выходим из функции
		err := os.Rename(fmt.Sprintf("./%s/%s", videoFilesDirectory, localVideoFileName), fmt.Sprintf("./%s/%s", resultFilesDirectory, localVideoFileName))
		if err != nil {
			return "", "", errors.New("Внутренняя ошибка сервера")
		}
		return localVideoFileName, "Ошибка при поиске аудиофайла в хранилище", nil
	}
	audioFileName = video.SessionId + ".mp3"

	log.Print("audioFileName", audioFileName)

	audioFile, err = os.Create(fmt.Sprintf("%s/%s", audioFilesDirectory, audioFileName))
	if err != nil {
		log.Error().Err(err)
		//перемещаем видео в папку resultVideos и выходим из функции
		err := os.Rename(fmt.Sprintf("./%s/%s", videoFilesDirectory, localVideoFileName), fmt.Sprintf("./%s/%s", resultFilesDirectory, localVideoFileName))
		if err != nil {
			return "", "", errors.New("Внутренняя ошибка сервера")
		}
		return localVideoFileName, "Ошибка при обработке аудиофайла", nil
	}
	defer audioFile.Close()

	buffer = make([]byte, 4096)
	for {
		n, err := info.Body.Read(buffer)
		if err != nil {
			break
		}
		audioFile.Write(buffer[:n])
	}

	fmt.Println("Audio downloaded successfully!")

	//Cлейка
	//ffmpeg -i VideoFiles/{video_name}.mkv  -i  AudioFiles/{audio_name}.mp3 -acodec copy -vcodec copy  VideoFiles/ResultVideos/{video_name}.mkv

	v := ffmpeg_go.Input(videoFilesDirectory + "/" + localVideoFileName)
	au := ffmpeg_go.Input(audioFilesDirectory + "/" + audioFileName)
	err = ffmpeg_go.Output([]*ffmpeg_go.Stream{v, au}, fmt.Sprintf("%s/%s", resultFilesDirectory, localVideoFileName), ffmpeg_go.KwArgs{"acodec": "copy", "vcodec": "copy"}).OverWriteOutput().Run()

	if err != nil {
		log.Printf("Не удалось склеить файл: %s\n", err.Error())
		//перемещаем видео в папку resultVideos и выходим из функции
		err := os.Rename(fmt.Sprintf("./%s/%s", videoFilesDirectory, localVideoFileName), fmt.Sprintf("./%s/%s", resultFilesDirectory, localVideoFileName))
		if err != nil {
			return "", "", errors.New("Внутренняя ошибка сервера")
		}
		return localVideoFileName, "Ошибка при обработке файла - не удалось склеить аудио с видеофайлом", nil
	}

	return localVideoFileName, "", nil
}

func (us *ScreenRecorderUseCases) DeleteCashe(ctx context.Context) error {

	var (
		audioFilesDirectory  string = "tmp/audioFiles"
		videoFilesDirectory  string = "tmp/videoFiles"
		resultFilesDirectory string = "tmp/resultVideos"
	)

	//Удалим видео, созданные более суток назад
	files, err := ioutil.ReadDir(resultFilesDirectory)
	if err != nil {
		log.Error().Err(err)
	}
	for _, file := range files {
		if file.ModTime().Unix() > time.Now().Add(24*time.Hour).Unix() {
			err := os.Remove(fmt.Sprintf("%s/%s", resultFilesDirectory, file.Name()))
			if err != nil {
				log.Error().Err(err)
				return err
			}
		}
	}

	//Удалим директорию videoFiles
	err = os.RemoveAll(fmt.Sprintf("%s", videoFilesDirectory))
	if err != nil {
		log.Error().Err(err)
		return err
	}

	//Удалим директорию audioFiles
	err = os.RemoveAll(fmt.Sprintf("%s", audioFilesDirectory))
	if err != nil {
		log.Error().Err(err)
		return err
	}

	//Cоздадим директорию videoFiles заново
	err = os.Mkdir(fmt.Sprintf("%s", videoFilesDirectory), os.FileMode(0522))
	if err != nil {
		log.Error().Err(err)
		return err
	}

	//Cоздадим директорию audioFiles заново
	err = os.Mkdir(fmt.Sprintf("%s", audioFilesDirectory), os.FileMode(0522))
	if err != nil {
		log.Error().Err(err)
		return err
	}

	return nil
}

func (us *ScreenRecorderUseCases) UpdatePostgresProjects(ctx context.Context) error {

	//Забрать pojects из Oracle

	projectsIdName, err := us.storeOra.ListProjectsIdName(ctx)
	if err != nil {
		return fmt.Errorf("ScreenRecorderUseCases - UpdatePostgresProjects - us.storeOra.ListProjectsIdName: %w", err)
	}

	//Обновить projects.name у уже существующих записей, а также вставить проекты, которых еще нет в базе.

	for _, pr := range projectsIdName {
		project := models.Project{
			UUID: pr.Id,
			Name: pr.Name,
		}
		err = us.storePg.AddProjectsFromOracle(ctx, project)
		if err != nil {
			return fmt.Errorf("ScreenRecorderUseCases - UpdatePostgresProjects - us.storePg.AddProjectsFromOracle: %w", err)
		}
	}

	return nil
}

func (us *ScreenRecorderUseCases) AddVideo(ctx context.Context, video models.Video) error {

	_, err := us.storePg.GetProject(ctx, video.ProjectUUID)
	if err != nil {
		switch {
		//Если проекта нет
		case strings.Contains(err.Error(), "Project does not exist"):
			p := models.Project{
				UUID: video.ProjectUUID,
				Name: "Неизвестный проект",
			}
			err = us.storePg.AddProject(ctx, p)
			if err != nil {
				return fmt.Errorf("ScreenRecorderUseCases - ListVideos - us.storePg.ListVideos: %w", err)
			}
		default:
			return fmt.Errorf("ScreenRecorderUseCases - ListVideos - us.storePg.ListVideos: %w", err)
		}
	}

	timeString := video.CreatedAt.Format("2006.01.02 15:04:05")
	video.CreatedAt, _ = time.Parse("2006.01.02 15:04:05", timeString)

	err = us.storePg.AddVideo(ctx, video)
	if err != nil {
		return fmt.Errorf("ScreenRecorderUseCases - ListVideos - us.storePg.ListVideos: %w", err)
	}

	return nil
}

func (us *ScreenRecorderUseCases) ListVideos(ctx context.Context, searchValue []models.SearchValue, orderValue []models.OrderValue, offset models.Offset) ([]models.ListVideos, error) {
	videos, err := us.storePg.ListVideos(ctx, searchValue, orderValue, offset)
	if err != nil {
		return nil, fmt.Errorf("ScreenRecorderUseCases - ListVideos - us.storePg.ListVideos: %w", err)
	}
	return videos, nil
}

func (us *ScreenRecorderUseCases) CountVideos(ctx context.Context, searchValue []models.SearchValue, orderValue []models.OrderValue, offset models.Offset) (int, error) {
	count, err := us.storePg.CountVideos(ctx, searchValue, orderValue, offset)
	if err != nil {
		return 0, fmt.Errorf("ScreenRecorderUseCases - CountVideos - us.storePg.CountVideos: %w", err)
	}
	return count, nil
}
