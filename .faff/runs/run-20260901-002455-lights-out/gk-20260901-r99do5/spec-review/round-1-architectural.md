(exit 10)
budget: backend openai/unsloth/Qwen3.8-27B-NVFP4: timeout 900s × ~6 worst-case (~5400s) >= per-backend budget 267s (deadline 800s / 3 backends) — retries/truncation may be cut short; lower this backend's timeout or raise --deadline
budget: backend openai/deepseek/deepseek-v4-flash-0731: timeout 900s × ~6 worst-case (~5400s) >= per-backend budget 267s (deadline 800s / 3 backends) — retries/truncation may be cut short; lower this backend's timeout or raise --deadline
budget: backend gemini/models/gemma-4-31b-it: timeout 900s × ~6 worst-case (~5400s) >= per-backend budget 267s (deadline 800s / 3 backends) — retries/truncation may be cut short; lower this backend's timeout or raise --deadline
[chain] openai/unsloth/Qwen3.8-27B-NVFP4 slice 267s exhausted → advancing (exit 8)
[chain] openai/deepseek/deepseek-v4-flash-0731 malformed (empty content) → advancing (exit 10)
exhausted: gemini/models/gemma-4-31b-it failed (transport-failed: HTTP 429: [{
  "error": {
    "code": 429,
    "message": "You exceeded your current quota, please check your plan and billing details. For more information on this error, head to: https://ai.google.dev/gemini-) (exit 5)
adversarial chain exhausted (3 backends, none produced findings); terminal exit 10 → needs-human
