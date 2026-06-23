package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"
	"nvidiagpt/handlers"
)

type Config struct {
	Port            string
	DBHost          string
	DBPort          string
	DBUser          string
	DBPassword      string
	DBName          string
	RedisHost       string
	RedisPort       string
	NvidiaAPIKey    string
	NvidiaAPIURL    string
	NvidiaModel     string
	ModelCategories []handlers.ModelCategory
}

func LoadConfig() *Config {
	// Load .env file if present
	godotenv.Load(".env")

	get := func(key, fallback string) string {
		if v := os.Getenv(key); v != "" {
			return v
		}
		return fallback
	}

	cfg := &Config{
		Port:         get("PORT", "8089"),
		DBHost:       get("DB_HOST", "localhost"),
		DBPort:       get("DB_PORT", "5432"),
		DBUser:       get("DB_USER", "nvidiagpt"),
		DBPassword:   get("DB_PASSWORD", "nvidiagpt"),
		DBName:       get("DB_NAME", "nvidiagpt"),
		RedisHost:    get("REDIS_HOST", "localhost"),
		RedisPort:    get("REDIS_PORT", "6379"),
		NvidiaAPIKey: get("NVIDIA_API_KEY", ""),
		NvidiaAPIURL: get("NVIDIA_API_URL", "https://integrate.api.nvidia.com/v1/chat/completions"),
		NvidiaModel: get("NVIDIA_MODEL", "meta/llama-3.3-70b-instruct"),
		ModelCategories: []handlers.ModelCategory{
			{
				Name: "Frontier / General LLMs",
				Models: []string{
					"nvidia/nemotron-3-ultra-550b-a55b",
					"moonshotai/kimi-k2-instruct",
					"deepseek-ai/deepseek-r1",
					"deepseek-ai/deepseek-v3",
					"meta/llama-3.3-70b-instruct",
					"meta/llama-3.1-405b-instruct",
					"meta/llama-3.1-70b-instruct",
					"qwen/qwen3-235b-a22b",
					"qwen/qwen3-32b",
					"mistralai/mistral-large-2-instruct",
					"mistralai/mixtral-8x22b-instruct-v0.1",
					"google/gemma-3-27b-it",
					"google/gemma-3-12b-it",
					"minimax-ai/minimax-m25",
					"zai-org/glm-5.1",
				},
			},
			{
				Name: "Text / Chat / Reasoning Models",
				Models: []string{
					"upstage/solar-10.7b-instruct",
					"mistralai/mixtral-8x7b-instruct-v0.1",
					"mistralai/mistral-nemotron",
					"meta/llama-4-maverick-17b-128e-instruct",
					"nvidia/llama-3.3-nemotron-super-49b-v1",
					"microsoft/phi-4-mini-instruct",
					"meta/llama-3.3-70b-instruct",
					"meta/llama-3.2-3b-instruct",
					"meta/llama-3.2-1b-instruct",
					"abacusai/dracarys-llama-3.1-70b-instruct",
					"nvidia/nemotron-mini-4b-instruct",
					"google/gemma-2-2b-it",
					"meta/llama-3.1-70b-instruct",
					"meta/llama-3.1-8b-instruct",
					"qwen/qwen3.5-397b-a17b",
					"stepfun-ai/step-3.5-flash",
					"mistralai/mistral-large-3-675b-instruct-2512",
					"mistralai/ministral-14b-instruct-2512",
					"bytedance/seed-oss-36b-instruct",
					"openai/gpt-oss-20b",
					"openai/gpt-oss-120b",
					"sarvamai/sarvam-m",
					"google/gemma-3n-e4b-it",
					"google/gemma-3n-e2b-it",
					"minimaxai/minimax-m3",
					"google/diffusiongemma-26b-a4b-it",
					"nvidia/nemotron-3-ultra-550b-a55b",
					"stepfun-ai/step-3.7-flash",
					"moonshotai/kimi-k2.6",
					"mistralai/mistral-medium-3.5-128b",
					"deepseek-ai/deepseek-v4-flash",
					"deepseek-ai/deepseek-v4-pro",
					"z-ai/glm-5.1",
					"minimaxai/minimax-m2.7",
					"google/gemma-4-31b-it",
					"mistralai/mistral-small-4-119b-2603",
					"nvidia/nemotron-3-super-120b-a12b",
					"qwen/qwen3.5-122b-a10b",
				},
			},
			{
				Name: "Coding Models",
				Models: []string{
					"qwen/qwen2.5-coder-32b-instruct",
					"deepseek-ai/deepseek-coder-v2-instruct",
					"meta/codellama-70b-instruct",
				},
			},
			{
				Name: "Reasoning Models",
				Models: []string{
					"deepseek-ai/deepseek-r1",
					"qwen/qwen3-235b-a22b",
					"moonshotai/kimi-k2-instruct",
					"nvidia/nemotron-3-ultra-550b-a55b",
					"nvidia/llama-3.3-nemotron-super-49b-v1",
					"openai/gpt-oss-120b",
					"openai/gpt-oss-20b",
					"stepfun-ai/step-3.5-flash",
					"google/gemma-4-31b-it",
					"mistralai/mistral-small-4-119b-2603",
				},
			},
			{
				Name: "Vision / Multimodal",
				Models: []string{
					"nvidia/nemotron-ocr-v1",
					"nvidia/nemoretriever-ocr",
					"google/paligemma",
					"nvidia/llama-3.1-nemotron-nano-vl-8b-v1",
					"microsoft/phi-4-multimodal-instruct",
					"meta/llama-3.2-11b-vision-instruct",
					"meta/llama-3.2-90b-vision-instruct",
					"nvidia/nemotron-nano-12b-v2-vl",
					"nvidia/nemotron-3-nano-omni-30b-a3b-reasoning",
					"nvidia/cosmos3-nano-reasoner",
					"qwen/qwen-image",
				},
			},
			{
				Name: "Speech & Translation",
				Models: []string{
					"nvidia/nemotron-asr-streaming",
					"nvidia/riva-translate-1.6b",
					"nvidia/riva-translate-4b-instruct-v1.1",
					"nvidia/riva-translate-4b-instruct-v1_1",
					"nvidia/magpie-tts-zeroshot",
					"nvidia/studio-voice",
					"nvidia/background-noise-removal",
					"nvidia/nemotron-voicechat",
				},
			},
			{
				Name: "Embedding Models",
				Models: []string{
					"nvidia/nv-embed-v1",
					"nvidia/nv-embedcode-7b-v1",
					"meta/esm2-650m",
				},
			},
			{
				Name: "Reranking Models",
				Models: []string{
					"nvidia/rerank-qa-mistral-4b",
				},
			},
			{
				Name: "Safety / Moderation",
				Models: []string{
					"nvidia/nemotron-content-safety-reasoning-4b",
					"nvidia/nemotron-3.5-content-safety",
					"nvidia/nemotron-3-content-safety",
					"meta/llama-guard-4-12b",
					"nvidia/llama-3.1-nemotron-safety-guard-8b-v3",
					"nvidia/gliner-pii",
				},
			},
			{
				Name: "Healthcare",
				Models: []string{
					"nvidia/llama-3.1-nemotron-nano-8b-healthcare-text2sql-v1.0",
					"nvidia/llama-3.3-nemotron-super-49b-healthcare-text2sql-v",
				},
			},
			{
				Name: "Biology",
				Models: []string{
					"meta/esmfold",
					"meta/esm2-650m",
				},
			},
			{
				Name: "Video",
				Models: []string{
					"nvidia/cosmos-transfer1-7b",
					"nvidia/cosmos-transfer2.5-2b",
					"nvidia/cosmos3-nano",
					"nvidia/synthetic-video-detector",
					"nvidia/active-speaker-detection",
				},
			},
			{
				Name: "Autonomous Driving",
				Models: []string{
					"nvidia/sparsedrive",
					"nvidia/bevformer",
					"nvidia/streampetr",
				},
			},
		},
	}

	if cfg.NvidiaAPIKey == "" {
		log.Println("WARNING: NVIDIA_API_KEY is not set. Set it in backend/.env")
	}

	return cfg
}
