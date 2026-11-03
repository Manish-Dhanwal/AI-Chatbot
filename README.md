# Banking Chatbot — learning project

This project is a small banking chatbot that we are growing step by step to learn AI and Agentic AI. Go stays the main backend language. We use local Ollama models so we can learn without a per-request API bill.

The bank name and policies in this project are fictional. Do not enter real account or personal information.

## Manual checks

Use [EVALUATION.md](EVALUATION.md) to check a direct question, a reworded question, and a question the FAQ cannot answer.

## Learning Q&A

Questions, doubts, simple explanations, and resolutions are recorded in [LEARNING_LOG.md](LEARNING_LOG.md). The README remains the project overview and decision log.

## What the app does today

A customer asks a question in the browser. Go uses a conversation ID to find recent messages, searches a small FAQ, and gives the recent messages and matching FAQ to a local language model. It returns the answer and FAQ source to the browser.

```text
Browser chat
  → Go POST /chat handler
  → Go loads recent messages for this conversation ID
  → Assistant service searches the FAQ by meaning
  → Ollama gpt-oss:20b uses the recent messages and selected FAQ to write an answer
  → Go returns the answer and FAQ source to the browser
```

A second local model, `nomic-embed-text`, helps with the FAQ search. It turns text into number lists that Go compares. It does not write answers.

## Run it

Start Ollama and make sure both models are installed:

```bash
ollama list
```

The expected models are `gpt-oss:20b` and `nomic-embed-text`. Then, from this project folder, run:

```bash
go run ./cmd/server
```

Open <http://localhost:8080> in a browser. The Go app reads `knowledge/bank_faq.txt` and creates the FAQ embeddings when it starts. Restart the app after changing that file.

To see which FAQ entry was selected and the model's reply in the terminal:

```bash
LOG_RETRIEVAL=true LOG_AI_RESPONSE=true go run ./cmd/server
```

Keep those logs off when using real or private information.

To run the focused retrieval unit test:

```bash
go test ./internal/assistant
```

These tests check Go's selection and cutoff logic with small example number lists. They verify that the assistant skips the model when no FAQ matches and sends the selected FAQ plus the original question when one does match. They do not check whether Ollama creates useful embeddings; the manual checks in [EVALUATION.md](EVALUATION.md) cover real questions through the running app.

## Project layout

| File or folder | Purpose |
|---|---|
| `cmd/server/main.go` | Loads the FAQ, creates the Ollama client and assistant service, and starts the Go web server. |
| `internal/httpapi/chat_handler.go` | Handles `POST /chat`, checks the request, calls the assistant, and returns JSON. |
| `internal/assistant/service.go` | Holds the banking assistant's instructions, prepares FAQ entries at startup, retrieves an entry for each question, asks the model to answer, and returns the source FAQ. |
| `internal/assistant/memory.go` | Keeps the last ten question-and-reply pairs for each conversation in Go memory. |
| `internal/assistant/retrieval.go` | Compares question and FAQ embeddings using cosine similarity. |
| `internal/assistant/retrieval_test.go` | Checks that Go picks a close FAQ and rejects a match below the cutoff, without needing Ollama. |
| `internal/assistant/service_test.go` | Checks that the assistant calls the model only when an FAQ matches. |
| `internal/ai/model.go` | Defines the small interfaces and message type used between the assistant and AI client. |
| `internal/ai/ollama.go` | Sends chat and embedding requests to the local Ollama server and reads its responses. |
| `knowledge/bank_faq.txt` | Fictional bank facts used as the source of truth for this demo. |
| `web/index.html` | Chat page, styles, and browser code that calls `/chat`. |
| `web/embed.go` | Embeds the page into the Go app so one server can serve both the UI and API. |

## How one question flows through the app

1. The browser sends JSON like `{"conversation_id":"...","message":"I need a fresh bank card urgently"}` to `POST /chat`.
2. The Go HTTP handler checks the request and passes the conversation ID and message to the assistant service.
3. The assistant loads recent messages for that ID from the in-memory conversation store.
4. At startup, the assistant has already asked `nomic-embed-text` to turn each FAQ entry into a number list. These lists stay in memory.
5. For each new question, the assistant asks `nomic-embed-text` for a number list. If this is a follow-up, it combines the last two conversation turns with the new question so FAQ search has enough context. Go selects the closest FAQ if its score is at least `0.60`.
6. If every entry scores lower, Go returns a fixed “FAQ does not specify” reply and saves that turn. It does not ask the LLM to answer without a matching FAQ.
7. If an entry matches, the assistant sends the system instruction, recent conversation messages, the FAQ entry, and the current question to `gpt-oss:20b`.
8. The assistant saves the question and reply, keeping only the latest ten pairs for that conversation.
9. Go returns the reply and FAQ source as JSON. The chat page lets the user expand “Show FAQ source.”

The number lists are only for finding relevant FAQ text. They are not sent to the answer-writing model.

## What we have learned and why we added each piece

### 1. A simple Go chat API and browser UI

**Need:** We needed a working starting point before adding AI.

**What we added:** `POST /chat` accepts a message and returns a reply. A small web page lets us use it in a browser.

**Why this first:** It gave us a clear request path to keep as we add AI features. The page and API use one Go server, so the browser does not need a separate cross-origin setup.

**Trade-off:** The first reply was fixed text. That was enough to prove the API/UI connection, but it could not answer questions.

### 2. One LLM call

**Need:** A fixed reply cannot respond to different customer messages.

**What we added:** A Go client sends the conversation to a model and returns its generated text.

**Provider choice:** We briefly tried Gemini, which returned a temporary high-demand error. We then tried the OpenAI API, but API usage is billed separately from ChatGPT. We switched to local Ollama to keep this learning project free of per-request API charges. The final code uses Ollama, not Gemini or OpenAI.

**Trade-off:** A local model uses computer memory and may be slower. It also does not have the reliability or capacity of a managed production model service.

### 3. Assistant instructions

**Need:** The model initially received only the customer's question. It did not know its role or limits in our demo.

**What we added:** A system instruction tells the model to act as a banking assistant, use the supplied FAQ, avoid making up facts, and not claim it can take banking actions.

**Trade-off:** Instructions guide the model; they do not guarantee it will obey. We saw the model add a delivery estimate that was not in the FAQ. That is why prompts alone are not a production security control.

### 4. A small fictional FAQ

**Need:** A model has general learned knowledge, but it does not know Northstar Bank's fictional policies.

**What we added:** `knowledge/bank_faq.txt` gives it approved facts, such as the replacement card fees.

**First approach:** We sent the whole short FAQ with every question. This is the simplest way to show grounding: the model gets source information with the question.

**Trade-off:** Sending all documents is fine for a tiny FAQ. With many or long documents it wastes space and may include unrelated information.

### 5. Keyword search, then meaning-based search

**Need:** We wanted Go to send only the FAQ entry relevant to a question.

**First approach:** Go counted shared words. It was easy to understand, but it missed some questions with different wording.

**Next approach:** We installed `nomic-embed-text`. It makes number lists that tend to be close for text with related meanings. Go compares those lists and selects the closest FAQ entry. This is a basic semantic retrieval step.

**Why keep the FAQ vectors in memory:** We have only four short entries. For this size, a database would add work without helping the lesson. The vectors are created once at startup and reused for questions.

**Trade-offs:** Similarity search is an estimate, not human understanding. The `0.60` minimum score is an initial cutoff based on our current examples, not a universal rule. It accepts the card paraphrase (`0.652`) and rejects the unrelated closing-time question (`0.537`). More examples may show that it needs adjustment. The model can still add unsupported details after retrieval.

### 6. Separate assistant behavior from Ollama communication

**Need:** Banking instructions and FAQ search describe our assistant. Calling Ollama is a separate technical job.

**What we did:** `internal/assistant` prepares the assistant's instructions and context. `internal/ai` handles communication with Ollama. The HTTP handler works with a small interface, not Ollama-specific details.

**Why:** This keeps each part focused and makes it easier to change the model provider later.

## Simple glossary

- **LLM:** A model that reads text and generates text. `gpt-oss:20b` writes the chatbot's answer.
- **System instruction:** Guidance about the assistant's role and behavior.
- **Grounding:** Giving the model trusted facts to use in its answer.
- **Embedding:** A list of numbers made from text. Similar meanings tend to produce nearby lists.
- **Retrieval:** Searching for the relevant FAQ entry before asking the answer-writing model.
- **RAG:** Retrieval-Augmented Generation: retrieve useful information, give it to the model, and generate an answer from it. Our current FAQ search is a small RAG example.
- **Cosine similarity:** The calculation Go uses to estimate how close two embedding lists are. The app logs the score when `LOG_RETRIEVAL=true`.

## Current limits and next learning steps

This is a learning demo, not a real banking service. It has fictional FAQ data and no real accounts, authentication, persistent database, or banking tools. Conversation memory is temporary: it lives in Go memory and is lost when the server restarts. The model cannot perform transactions, though it may still phrase an answer as if it can; generated text must not be treated as authorization.

Next we will improve one thing at a time. Later we can move conversation history to Postgres, add safe Go tools, and build agent workflows. Python can be introduced when we reach AI orchestration; Go remains the backend.

## Decision log

We will add a short entry here whenever we introduce a new part:

- **Problem:** What was missing or failing?
- **Change:** What did we add?
- **Reason:** Why this was the smallest useful next step?
- **Trade-off:** What limitation did the choice bring?
- **Check:** What example shows whether it helped?

### Current retrieval decision

- **Problem:** Word matching missed questions phrased differently from the FAQ. Then an unrelated closing-time question was matched to the domestic-transfer FAQ because the cutoff was too low.
- **Change:** Use local embeddings and require a score of at least `0.60`; lower scores make Go return a fixed “FAQ does not specify” reply without calling the LLM.
- **Reason:** It keeps our observed card paraphrase (`0.652`) and rejects the unrelated match (`0.537`).

### Current conversation memory decision

- **Problem:** Each request used to be independent, so the assistant could not understand a follow-up question.
- **Change:** The browser sends a conversation ID. Go keeps the last ten question-and-reply pairs for each ID, uses the last two turns to give follow-up FAQ search context, and includes the history in the next matching LLM request.
- **Reason:** This teaches conversation memory without adding a database before we understand the flow.
- **Trade-off:** Memory is lost on server restart and is not shared between multiple server instances. Using recent turns can mix topics if the customer changes subjects quickly. We will replace this store with Postgres next.
- **Check:** Ask about an expedited replacement, then ask “What about the standard replacement?” in the same browser tab. The next answer should use the earlier context and show the matching FAQ source.
- **Trade-off:** These are only a few examples. Similarity scores are not probabilities, and the cutoff may reject a valid question or accept an unrelated one in other cases.
- **Check:** Use [EVALUATION.md](EVALUATION.md) to check direct questions, reworded questions, and a question with no answer in the FAQ. The closing-time question should log “no relevant entry.”
