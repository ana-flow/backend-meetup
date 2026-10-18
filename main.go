package main

import (
	"fmt"      // используется для форматирования строк (например, сборки URL-ссылок)
	"net/http" // содержит константы HTTP-статусов (200 OK, 404 Not Found)
	"strconv"  // strconv превращает строки в числа и наоборот (для параметров запроса ?id=1)

	"github.com/gin-gonic/gin" // веб-фреймворк для языка Go
)

// MinioBaseURL - константа с публичным адресом нашего бакета в MinIO.
// Используется для формирования прямых ссылок на медиафайлы.
const MinioBaseURL = "http://localhost:9000/meetup-bucket"

// Speaker - структура (struct) в Go
// Описывает сущность "Спикер по ИИ" по предметной области.
// Все поля пишутся с большой буквы, чтобы Go разрешил их чтение снаружи (публичные поля).
type Speaker struct {
	ID              int    `json:"id"`
	FullName        string `json:"fullName"`
	TopicTitle      string `json:"topicTitle"`
	Description     string `json:"description"`     // Единое поле: Описание спикера (экспертиза, био)
	BirthDate       string `json:"birthDate"`       // Поле 1 по теме: Дата рождения "ДД.ММ.ГГГГ"
	ExperienceYears int    `json:"experienceYears"` // Поле 2 по теме: Опыт работы в годах
	SpeakerStatus   string `json:"speakerStatus"`   // Статус: "DRAFT", "PUBLISHED", "DELETED"
	VideoKey        string `json:"videoKey"`        // Ключ видеофайла в MinIO
	AvatarKey       string `json:"avatarKey"`       // Ключ изображения в MinIO
	Likes           []int  `json:"likes"`           // Срез ID пользователей, поставивших лайки
}

// VideoURL - метод структуры Speaker. Возвращает полную ссылку на видео из MinIO.
func (s Speaker) VideoURL() string {
	return fmt.Sprintf("%s/%s", MinioBaseURL, s.VideoKey)
}

// AvatarURL - метод структуры Speaker. Возвращает полную ссылку на аватарку из MinIO.
func (s Speaker) AvatarURL() string {
	return fmt.Sprintf("%s/%s", MinioBaseURL, s.AvatarKey)
}

// LikesCount - вычисляемый метод возвращает количество лайков через len()
func (s Speaker) LikesCount() int {
	return len(s.Likes)
}

// Speakers - массив (срез/slice в Go) с тестовыми данными.
// Данные хранятся прямо в памяти сервера без использования базы данных.
var Speakers = []Speaker{
	{
		ID:              1,
		FullName:        "Алекс Мерсер",
		TopicTitle:      "LLM в продакшене",
		Description:     "Эксперт по оптимизации и локальному запуску LLM с 7-летним опытом разработки ИИ-систем.",
		BirthDate:       "14.05.1991",
		ExperienceYears: 7,
		SpeakerStatus:   "PUBLISHED",
		VideoKey:        "Alex_Mercer.mp4",
		AvatarKey:       "Alex_Mercer_avatar.jpg",
		Likes:           []int{0, 1, 2, 3, 4, 5, 6},
	},
	{
		ID:              2,
		FullName:        "Елена Ростова",
		TopicTitle:      "Автономные AI-агенты для E-commerce",
		Description:     "Архитектор мультиагентных систем для автоматизации e-commerce процессов и обработки заказов.",
		BirthDate:       "22.11.1995",
		ExperienceYears: 5,
		SpeakerStatus:   "PUBLISHED",
		VideoKey:        "Elena_Rostova.mp4",
		AvatarKey:       "Elena_Rostova_avatar.jpg",
		Likes:           []int{0, 5, 9},
	},
	{
		ID:              3,
		FullName:        "Игорь Сидоров",
		TopicTitle:      "Computer Vision в ритейле",
		Description:     "Ведущий инженер Computer Vision, специалист по внедрению нейросетей YOLO на производстве.",
		BirthDate:       "03.08.1988",
		ExperienceYears: 10,
		SpeakerStatus:   "PUBLISHED",
		VideoKey:        "Igor_Sidorov.mp4",
		AvatarKey:       "Igor_Sidorov_avatar.jpg",
		Likes:           []int{1, 2, 3, 4, 5, 6, 7, 8, 9},
	},
	{
		ID:              4,
		FullName:        "Сара Чен",
		TopicTitle:      "AI Safety & Prompt Injection",
		Description:     "Исследователь безопасности RAG-систем и киберзащиты корпоративных баз знаний от промпт-атак.",
		BirthDate:       "19.02.1994",
		ExperienceYears: 6,
		SpeakerStatus:   "DRAFT",
		VideoKey:        "Sarah_Chen.mp4",
		AvatarKey:       "Sarah_Chen_avatar.jpg",
		Likes:           []int{},
	},
	{
		ID:              5,
		FullName:        "Удаленный Спикер",
		TopicTitle:      "Устаревший доклад",
		Description:     "Этот спикер удален и не выводится в интерфейсе.",
		BirthDate:       "01.01.2000",
		ExperienceYears: 1,
		SpeakerStatus:   "DELETED",
		VideoKey:        "Alex_Mercer.mp4",
		AvatarKey:       "Alex_Mercer_avatar.jpg",
		Likes:           []int{},
	},
}

func main() {
	// Создаем экземпляра роутера (r - router) Gin с логированием по умолчанию
	r := gin.Default()

	// Загружаем HTML-шаблоны из папки templates
	r.LoadHTMLGlob("templates/*")

	// Раздаем статические CSS-файлы из папки static по URL-адресу /static/...
	r.Static("/static", "./static")

	// GET /feed (Экран "Лента")
	// Поддерживает query-параметры: ?id=1 и ?next=true
	r.GET("/feed", func(c *gin.Context) {
		idStr := c.Query("id")     // Читаем параметр id
		nextStr := c.Query("next") // Читаем параметр next

		// Отбираем только опубликованных спикеров (SpeakerStatus == "PUBLISHED")
		var published []Speaker
		for _, s := range Speakers {
			if s.SpeakerStatus == "PUBLISHED" {
				published = append(published, s)
			}
		}

		if len(published) == 0 {
			c.String(http.StatusNotFound, "Опубликованные спикеры не найдены")
			return
		}

		// По умолчанию выбираем первого спикера
		selectedSpeaker := published[0]

		if idStr != "" {
			currentID, _ := strconv.Atoi(idStr)

			// Ищем текущего спикера по ID
			currentIndex := -1
			for i, s := range published {
				if s.ID == currentID {
					currentIndex = i
					break
				}
			}

			if currentIndex != -1 {
				if nextStr == "true" {
					// Если указан next=true, берем следующего элемента (с зацикливанием по кругу)
					nextIndex := (currentIndex + 1) % len(published)
					selectedSpeaker = published[nextIndex]
				} else {
					selectedSpeaker = published[currentIndex]
				}
			}
		}

		// Рендерим HTML-шаблон feed.html и передаем в него данные
		c.HTML(http.StatusOK, "feed.html", gin.H{
			"speaker":    selectedSpeaker,
			"videoUrl":   selectedSpeaker.VideoURL(),
			"likesCount": selectedSpeaker.LikesCount(),
		})
	})

	// GET /draft (Экран "Добавление")
	// Отображает услугу со статусом DRAFT
	r.GET("/draft", func(c *gin.Context) {
		var draftSpeaker *Speaker

		// Ищем первого спикера со статусом DRAFT
		for _, s := range Speakers {
			if s.SpeakerStatus == "DRAFT" {
				draftSpeaker = &s
				break
			}
		}

		if draftSpeaker == nil {
			c.String(http.StatusNotFound, "Черновик не найден")
			return
		}

		// Рендерим HTML-шаблон draft.html
		c.HTML(http.StatusOK, "draft.html", gin.H{
			"speaker":   draftSpeaker,
			"videoUrl":  draftSpeaker.VideoURL(),
			"avatarUrl": draftSpeaker.AvatarURL(),
		})
	})

	// GET /catalog (Экран "Плитка")
	// Фильтрует список по полю опыта работы (min_experience)
	r.GET("/catalog", func(c *gin.Context) {
		minExperienceStr := c.Query("min_experience")

		minExp := 0
		if minExperienceStr != "" {
			minExp, _ = strconv.Atoi(minExperienceStr)
		}

		// Собираем список карточек для отображения
		var result []map[string]interface{}

		for _, s := range Speakers {
			// Исключаем статусы DELETED и DRAFT
			if s.SpeakerStatus != "DELETED" && s.SpeakerStatus != "DRAFT" {
				if s.ExperienceYears >= minExp {
					result = append(result, map[string]interface{}{
						"speaker":    s,
						"videoUrl":   s.VideoURL(),
						"avatarUrl":  s.AvatarURL(),
						"likesCount": s.LikesCount(),
					})
				}
			}
		}

		// Рендерим HTML-шаблон catalog.html
		c.HTML(http.StatusOK, "catalog.html", gin.H{
			"speakers":      result,
			"minExperience": minExperienceStr, // Сохраняем введенное значение фильтра
		})
	})

	// Запускаем веб-сервер на порту 3000
	r.Run(":3000")
}
