# Learning Q&A Log

This file records the questions that were confusing, the simple explanation, and what we decided to do. New learning questions will be added here.

## 2026-11-03 — Why search the FAQ instead of sending it all?

**Question / doubt**

If we send the FAQ and the customer question to the LLM, why does Go search for a matching FAQ first? What is the LLM doing then?

**Simple answer**

At first, we did send the whole small FAQ with the question. That works for four short facts, and the LLM can choose what seems relevant.

Now Go searches first. It sends the closest FAQ fact and the customer's question to the LLM. Go is finding the source fact; the LLM is writing the final answer in natural language.

**Why add search?**

When a knowledge base has hundreds or thousands of entries, sending every entry with every question uses more space and includes unrelated facts. Search lets us send only the most useful fact. That keeps the request smaller and helps the model focus.

**Trade-off**

Search adds work and can choose the wrong fact. For our four-entry FAQ, sending everything is simpler. We added search to learn the pattern that becomes useful as the knowledge base grows. Our similarity score limit is only a starting point and needs examples to tune it.

**Resolution**

The current flow is:

1. Go searches the FAQ by meaning.
2. Go sends the selected FAQ fact and the original question to the LLM.
3. The LLM writes the answer.

For a very simple question, Go could return the FAQ fact directly and skip the LLM. We keep the LLM here to learn how retrieved information can guide generated answers.

## 2026-11-03 — What does `nomic-embed-text` do?

**Question / doubt**

How can it match different words that mean something similar?

**Simple answer**

It turns a sentence into a list of numbers. Sentences about similar things tend to get number lists that are close together. Go compares the lists to find the closest FAQ entry.

For example, “fresh bank card urgently” can be close to “expedited debit card replacement.” The number lists are only for search; they are not sent to the answer-writing model.

**Resolution**

`nomic-embed-text` searches. `gpt-oss:20b` writes the answer.

## 2026-11-03 — Why did the model add a delivery time?

**Question / doubt**

The FAQ did not say how quickly the card would arrive, but the model added a delivery estimate and offered to start a request.

**Simple answer**

Search found a useful FAQ fact, but that does not force the LLM to use only that fact. It may add details from its general language patterns.

**Resolution**

We made the fictional FAQ clear about what it does and does not say, and made the assistant instructions more specific. The model now gives the fee and says the FAQ does not state the delivery time. Instructions help, but they are not a security guarantee. Go must enforce real actions and permissions later.

## 2026-11-03 — Why use a system instruction?

**Question / doubt**

Why send an instruction with the customer's message?

**Simple answer**

The customer message says what the customer wants. The system instruction tells the model what role it has and how it should behave in this app. We send both to Ollama.

**Trade-off**

The system instruction guides the model but cannot guarantee its behavior. Important permissions and banking rules must be enforced by Go code, not only by an instruction.

## 2026-11-03 — What does the current search do?

**Question / doubt**

Does Go extract words from the question, match them to the FAQ, and send the matching FAQ with the question to the LLM?

**Simple answer**

That was our first search version. It counted matching words, but we replaced it because different wording could miss the right FAQ.

The current version turns the question and each FAQ entry into number lists that represent meaning. Go compares those lists, picks the closest FAQ entry, and sends that entry plus the original question and system instruction to the LLM. The LLM writes the response.

**Resolution**

Current flow: question → meaning-based search → one FAQ entry → LLM answer. The whole FAQ and the word maps are not sent to the LLM.

## 2026-11-03 — Are we using GloVe embeddings?

**Question / doubt**

Are we using word embeddings to find the closest FAQ when the wording differs?

**Simple answer**

Yes, we use embeddings to find a close FAQ. The model in this project is `nomic-embed-text`, not GloVe. Nomic turns the whole question and each FAQ entry into number lists. Go compares those lists and selects the closest FAQ entry.

The selected FAQ, original question, and system instruction then go to `gpt-oss:20b`, which writes the response.

**Resolution**

GloVe is a different embedding method. We use Nomic text embeddings for this project.

## 2026-11-03 — Why add a Go retrieval test?

**Goal**

Check the FAQ selection rule automatically when we change the Go code.

**Simple answer**

The test gives Go small example number lists and checks two rules: select the closest FAQ above `0.60`, and reject it below `0.60`.

**Why**

It is fast and repeatable, and it does not need Ollama. It catches changes that break Go's selection and cutoff decisions.

**Trade-off**

The test uses made-up number lists. It checks Go's decision, but not whether Nomic creates useful number lists for real customer questions. We still need the manual app checks for that.

## 2026-11-03 — Does Go skip the LLM when no FAQ matches?

**Goal**

Make sure an unrelated FAQ question does not get sent to the answer-writing model.

**What we added**

A service test uses a fake embedding model and a fake answer model. It checks that a below-cutoff question gets Go's fixed reply with zero answer-model calls, while a matching question sends the selected FAQ and original question to the model.

**Why**

The earlier retrieval test checked only the search result. This test checks the next step too: whether the assistant follows that result safely.

**Trade-off**

The fakes make the test repeatable and do not need Ollama, but they do not test real embedding quality or model behavior.

## 2026-11-03 — Show which FAQ supports an answer

**Need**

The chat showed the model's reply, but not which bank FAQ Go supplied to the model.

**What we added**

The assistant result now carries both the reply and the selected FAQ source. The API returns both, and the chat page has a “Show FAQ source” control under answers with a match. If no FAQ matches, the source is empty and the page shows only Go's fixed reply.

**Why**

Showing the source helps a user or developer check which bank fact supports the reply and notice a bad match.

**Trade-off**

Showing a source helps people inspect grounding, but does not prove the model followed it or that the FAQ itself is correct.

## 2026-11-03 — Keep recent conversation memory in Go

**Need**

Each question was handled on its own, so follow-up questions could lose context.

**What we added**

The browser creates a conversation ID and sends it with each question. Go stores up to ten recent question-and-reply pairs for that ID, then includes them in later LLM requests that have a matching FAQ. Go also saves its fixed no-FAQ reply as a turn.

**Why**

The LLM can use earlier messages to understand a follow-up such as “What about the standard replacement?”

**Trade-off**

This store is protected for concurrent requests, but it is temporary and local to one server process. Restarting Go loses all conversations. The next storage step can move the same store interface to Postgres.

## 2026-11-03 — Give FAQ search context for follow-up questions

**What happened**

After asking about a fresh bank card, “What about the standard option?” got Go's no-FAQ reply. The search looked only at the short follow-up, so it did not connect it to the earlier card question.

**What we changed**

Before searching, Go combines the last two conversation turns with the new follow-up. The LLM still receives the original conversation messages and the selected FAQ.

**Why**

Both retrieval and the LLM need enough context to understand a follow-up. Previously, only the LLM had conversation history, but Go searched before it called the LLM. Using only the last question also failed when the customer repeated an ambiguous follow-up, so we now keep two recent turns in the search text.

**Trade-off**

Combining recent turns is a simple first step. It may mix subjects if the customer changes topics; later we can learn query rewriting and more careful context selection.

**Check**

After including the last two turns in the search text, the customer repeated “What about the standard option?” and Go kept selecting the debit card replacement FAQ. The observed scores were above the `0.60` cutoff, and the replies showed that FAQ as their source.
