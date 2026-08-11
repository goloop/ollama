# ollama - довідник

Повний довідник пакета `ollama`: клієнт, спільна модель `goloop/ai`, чат
(інтерфейс і нативний), стрімінг, embeddings і моделі.

Англійська версія: **[DOC.md](DOC.md)**.

## Зміст

- [Ментальна модель](#ментальна-модель)
- [Створення клієнта](#створення-клієнта)
- [Generate і Stream](#generate-і-stream)
- [Структурований вивід](#структурований-вивід)
- [Пошук на боці провайдера](#пошук-на-боці-провайдера)
- [Нативний чат](#нативний-чат)
- [Інструменти й зображення](#інструменти-й-зображення)
- [Embeddings](#embeddings)
- [Моделі](#моделі)
- [Опції та помилки](#опції-та-помилки)

## Ментальна модель

`ollama.Client` реалізує `ai.Client` - провайдер-незалежний контракт із
`github.com/goloop/ai`. Спільні `Generate` і `Stream` покривають спільну основу
(чат із інструментами, зображеннями й стрімінгом), тож код проти інтерфейсу
працює з будь-яким провайдером - локальним чи хмарним.

Нативна частина - `ChatCompletion`/`ChatStream` через `/api/chat`, embeddings і
перелік встановлених моделей. Ollama стрімить JSON порядково (newline-delimited),
а не Server-Sent Events; драйвер це приховує.

```go
import (
	"github.com/goloop/ai"
	"github.com/goloop/ollama"
)
```

## Створення клієнта

```go
c := ollama.New("") // локальний сервер на http://localhost:11434

c = ollama.New(apiKey, ollama.WithBaseURL("https://ollama.example.com"))
```

Ключ необовʼязковий; коли заданий - надсилається як bearer-токен для
автентифікованих проксі перед сервером.

## Generate і Stream

```go
resp, err := c.Generate(ctx, &ai.Request{
	Model:    ollama.ModelLlama32,
	System:   "You are concise.",
	Messages: []ai.Message{ai.UserText("Name three primary colors.")},
})
resp.Text()
resp.ToolCalls()
resp.Usage
```

`Stream` повертає `iter.Seq2[ai.Chunk, error]`: текстові дельти чанками з `Text`,
виклик інструмента - чанком із `ToolCall`, фінальний чанк - `Done` і `Usage`.

```go
for chunk, err := range c.Stream(ctx, req) {
	if err != nil {
		return err
	}
	fmt.Print(chunk.Text)
}
```

Якщо стрім завершується до того, як Ollama позначить його завершеним, `Stream`
віддає `io.ErrUnexpectedEOF`, а не тихо повідомляє про завершену відповідь.

## Структурований вивід

`ai.Request.Format` лягає на власний поле `format` провайдера, тож запит на JSON
провайдер **дотримує**, а не просто «чує»:

```go
resp, err := c.Generate(ctx, &ai.Request{
	Model:    "the-model",
	Messages: []ai.Message{ai.UserText("Склади SEO-поля для цієї статті.")},
	Format: &ai.Format{
		Type:   ai.FormatJSONSchema,
		Name:   "seo",
		Schema: schema,
	},
})

var seo SEO
err = resp.JSON(&seo)
```

`ai.FormatJSON` надсилає голе слово `"json"`, а `ai.FormatJSONSchema` - саму
схему: жодної обгортки навколо неї немає.


`ai.Response.Format` дорівнює `ai.FormatNative`: цей провайдер дотримує кожну
форму, яку приймає. Які моделі підтримують схемний режим - справа провайдера;
таблиці можливостей тут немає, тож про непідтримувану пару скаже він сам.

## Нативний чат

Для опцій, специфічних для провайдера, будуйте `ChatRequest` і викликайте
`ChatCompletion` чи `ChatStream`:

```go
resp, err := c.ChatCompletion(ctx, &ollama.ChatRequest{
	Model:    ollama.ModelLlama32,
	Messages: []ollama.Message{{Role: "user", Content: "as JSON"}},
	Format:   json.RawMessage(`"json"`),
	Options:  &ollama.Options{NumPredict: 256},
})
```

`Options` мапиться на опції генерації Ollama (`temperature`, `top_p`,
`num_predict`, `stop`); `Format` задає структурований вивід.

## Інструменти й зображення

Інструменти й зображення використовують спільні типи `ai`: `ai.Tool`,
`ai.Image`, `ai.ToolResult`. Ollama зіставляє результати з викликами позиційно,
а не за ID, тож повернений `ai.ToolUse` несе імʼя функції як свій `ID`. Вбудовані
байти зображення надсилаються у base64 (зображення за URL сервер не завантажує).

## Embeddings

```go
vecs, err := c.Embed(ctx, ollama.ModelLlama32, "hello", "world")
```

## Моделі

```go
models, err := c.Models(ctx) // встановлені моделі (ендпоінт /api/tags)
models[0].Name               // "llama3.2:latest"
models[0].Details.ParameterSize
```

## Пошук на боці провайдера

`ai.Request.Hosted` отримує `ai.ErrNoHosted` ще до відправлення запиту.

Моделі тут виконуються на тій самій машині, що їх обслуговує, і в цього сервера
немає пошуку, який можна запустити. Немає ані ендпоїнта, куди йти, ані джерел,
які повертати.

Ця відмова - задокументована поведінка, а не мовчазна прогалина: відповідь, яку
дали без замовленого пошуку, виглядає точно так само, як та, що з пошуком, тож
голосна помилка - єдиний спосіб їх розрізнити. Якщо відповідь усе одно потрібна,
повторіть запит без `Hosted`.

## Опції та помилки

Опції: `WithBaseURL`, `WithHTTPClient`, `WithTimeout`, `WithMaxRetries`,
`WithHeader`.

Невдала відповідь стає `*ai.APIError` зі `Status`, `Message` і сирим тілом:

```go
var apiErr *ai.APIError
if errors.As(err, &apiErr) && apiErr.Status == http.StatusNotFound {
	// модель ще не завантажено
}
```

Запити без моделі чи повідомлень падають до мережі з `ai.ErrNoModel` або
`ai.ErrNoMessages`.
