# Chatbot check list

Use these three questions to see whether FAQ search and answers are behaving as expected. The exact wording of the reply can vary; check that it uses the right fact and does not invent details.

| Customer question | FAQ entry Go should select | A good reply should |
|---|---|---|
| What does an expedited debit card replacement cost? | Debit card replacement | Say it costs $15. Do not invent a delivery time or claim to place an order. |
| I need a fresh bank card urgently. | Debit card replacement | Find the same entry even though the words differ. Say expedited replacement costs $15 and delivery timing is not specified. Do not claim to place an order. |
| How much do you charge for each item that overdraws my account? | Overdraft fee | Say $20 per overdraft item, with a maximum of two fees per day. |
| How long does a domestic bank transfer usually take? | Domestic transfer | Say it usually completes within one business day. |
| What time does Northstar Bank close? | No entry | Log `no relevant entry`. The fixed reply should say the FAQ does not specify the hours. |

## How to use this list

1. Start the app with retrieval logging enabled: `LOG_RETRIEVAL=true go run ./cmd/server`.
2. Ask each question in the chat page.
3. Check the terminal log to see which FAQ entry was selected (or whether none was selected).
4. Compare the log and reply with the table. For the closing-time question, the terminal should say `no relevant entry`, and Go should return its fixed FAQ-missing reply without asking Ollama.

If a question with an answer in the FAQ logs `no relevant entry`, the cutoff may be too high or the question may need a better FAQ match. If an unrelated FAQ is selected, the cutoff may be too low. Write down the question, log, and reply so we can decide what to adjust.

## Results observed on 2026-11-03

| Question | Observed result | Outcome |
|---|---|---|
| What does an expedited debit card replacement cost? | Debit card replacement, score `0.881` | Passed; reply said $15. |
| I need a fresh bank card urgently. | Debit card replacement, score `0.652` | Passed; reply used the fee and did not invent delivery timing. |
| How much do you charge for each item that overdraws my account? | Overdraft fee, score `0.781` | Passed. |
| How long does a domestic bank transfer usually take? | Domestic transfer, score `0.890` | Passed; reply said usually one business day. |
| What time does Northstar Bank close? | No relevant entry, score `0.537` | Passed; Go rejected the match and used its fixed reply. |

## Conversation memory check observed on 2026-11-03

The customer asked about a fresh bank card, then asked “What about the standard option?” twice. Each logged search selected the debit card replacement FAQ. The logged scores ranged from `0.652` to `0.924`, all above the `0.60` cutoff. Both follow-up replies said the standard replacement is free and displayed the FAQ source.
