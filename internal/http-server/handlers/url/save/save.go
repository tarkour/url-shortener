package save

import (
	"errors"
	"io"
	"log/slog"
	"net/http"
	"url-shortener/internal/lib/api/response"
	"url-shortener/internal/lib/random"
	"url-shortener/internal/lib/sl"
	"url-shortener/internal/storage"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
)

// TODO: move to config if need
const (
	aliasLength = 6
)

type Request struct {
	URL   string `json:"url" validate:"required,url"` // validate checks if url(which is must have) correct
	Alias string `json:"alias,omitempty"`             // omitempty - if no value to give for json in  - it will be absence full. without omitempty parametr value will be empty, but parametr will be anyway
}

type Response struct {
	response.Response
	Alias string `json:"alias,omitempty"`
}

//go:generate go run github.com/vektra/mockery/v2@v2.28.2 --name=URLSaver

type URLSaver interface {
	SaveURL(urlToSave string, alias string) (int64, error)
}

func New(log *slog.Logger, urlSaver URLSaver) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.url.save.New" // op = operation

		log = log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		var req Request

		err := render.DecodeJSON(r.Body, &req) //parsing request

		if errors.Is(err, io.EOF) {
			// Такую ошибку встретим, если получили запрос с пустым телом.
			// Обработаем её отдельно
			log.Error("request body is empty")

			render.JSON(w, r, response.Error("empty request"))

			return
		}

		if err != nil {
			log.Error("failed to decode request body", sl.Err(err))

			render.JSON(w, r, response.Error("failed to decode request")) // does not interrupt

			return
		}

		log.Info("request body decoded", slog.Any("request", req))

		if err := validator.New().Struct(req); err != nil { // creating new validator and analyse struct (scruct Request)
			validateErr := err.(validator.ValidationErrors) // cast error to desired type

			log.Error("invalid request", sl.Err(err)) //log of error without  any modifications

			render.JSON(w, r, response.ValidateError(validateErr)) // forming a request to put there normal-looking(for human) log of error

			return
		}
		//TODO: case if generated alias already exists
		alias := req.Alias
		if alias == "" {
			alias = random.NewRandomString(aliasLength)
		}

		id, err := urlSaver.SaveURL(req.URL, alias)
		if errors.Is(err, storage.ErrURLExists) {
			log.Info("url already exists", slog.String("url", req.URL))

			render.JSON(w, r, response.Error("url already exists"))

			return
		}

		if err != nil {
			log.Error("failed to add url", sl.Err(err))

			render.JSON(w, r, response.Error("failed to add url"))

			return
		}

		log.Info("url added", slog.Int64("id", id)) //return succsesful answer

		responseOK(w, r, alias)

	}
}

func responseOK(w http.ResponseWriter, r *http.Request, alias string) {
	render.JSON(w, r, Response{
		Response: response.OK(),
		Alias:    alias,
	})
}
