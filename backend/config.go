package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"
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
	AvailableModels []string
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
		NvidiaModel:  get("NVIDIA_MODEL", "meta/llama-4-maverick-17b-128e-instruct"),
		AvailableModels: []string{
			// 01.AI
			"01-ai/yi-large",
			// Abacus.AI
			"abacusai/dracarys-llama-3.1-70b-instruct",
			// Adept
			"adept/fuyu-8b",
			// AI21 Labs
			"ai21labs/jamba-1.5-large-instruct",
			// AI Singapore
			"aisingapore/sea-lion-7b-instruct",
			// BAAI
			"baai/bge-m3",
			// BigCode
			"bigcode/starcoder2-15b",
			// ByteDance
			"bytedance/seed-oss-36b-instruct",
			// Databricks
			"databricks/dbrx-instruct",
			// DeepSeek AI
			"deepseek-ai/deepseek-coder-6.7b-instruct",
			"deepseek-ai/deepseek-v4-flash",
			"deepseek-ai/deepseek-v4-pro",
			// Google
			"google/codegemma-1.1-7b",
			"google/codegemma-7b",
			"google/deplot",
			"google/diffusiongemma-26b-a4b-it",
			"google/gemma-2-2b-it",
			"google/gemma-2b",
			"google/gemma-3-12b-it",
			"google/gemma-3-4b-it",
			"google/gemma-3n-e2b-it",
			"google/gemma-3n-e4b-it",
			"google/gemma-4-31b-it",
			"google/recurrentgemma-2b",
			// IBM
			"ibm/granite-3.0-3b-a800m-instruct",
			"ibm/granite-3.0-8b-instruct",
			"ibm/granite-34b-code-instruct",
			"ibm/granite-8b-code-instruct",
			// Meta
			"meta/codellama-70b",
			"meta/llama-3.1-70b-instruct",
			"meta/llama-3.1-8b-instruct",
			"meta/llama-3.2-11b-vision-instruct",
			"meta/llama-3.2-1b-instruct",
			"meta/llama-3.2-3b-instruct",
			"meta/llama-3.2-90b-vision-instruct",
			"meta/llama-3.3-70b-instruct",
			"meta/llama-4-maverick-17b-128e-instruct",
			"meta/llama-guard-4-12b",
			"meta/llama2-70b",
			// Microsoft
			"microsoft/kosmos-2",
			"microsoft/phi-3-vision-128k-instruct",
			"microsoft/phi-3.5-moe-instruct",
			"microsoft/phi-4-mini-instruct",
			"microsoft/phi-4-multimodal-instruct",
			// MiniMax AI
			"minimaxai/minimax-m2.7",
			"minimaxai/minimax-m3",
			// Mistral AI
			"mistralai/codestral-22b-instruct-v0.1",
			"mistralai/ministral-14b-instruct-2512",
			"mistralai/mistral-7b-instruct-v0.3",
			"mistralai/mistral-large",
			"mistralai/mistral-large-2-instruct",
			"mistralai/mistral-large-3-675b-instruct-2512",
			"mistralai/mistral-medium-3.5-128b",
			"mistralai/mistral-nemotron",
			"mistralai/mistral-small-4-119b-2603",
			"mistralai/mixtral-8x22b-v0.1",
			"mistralai/mixtral-8x7b-instruct-v0.1",
			// Moonshot AI
			"moonshotai/kimi-k2.6",
			// NV-Mistral AI
			"nv-mistralai/mistral-nemo-12b-instruct",
			// NVIDIA
			"nvidia/ai-synthetic-video-detector",
			"nvidia/cosmos-reason2-8b",
			"nvidia/embed-qa-4",
			"nvidia/gliner-pii",
			"nvidia/ising-calibration-1-35b-a3b",
			"nvidia/llama-3.1-nemoguard-8b-content-safety",
			"nvidia/llama-3.1-nemoguard-8b-topic-control",
			"nvidia/llama-3.1-nemotron-51b-instruct",
			"nvidia/llama-3.1-nemotron-70b-instruct",
			"nvidia/llama-3.1-nemotron-nano-8b-v1",
			"nvidia/llama-3.1-nemotron-nano-vl-8b-v1",
			"nvidia/llama-3.1-nemotron-safety-guard-8b-v3",
			"nvidia/llama-3.1-nemotron-ultra-253b-v1",
			"nvidia/llama-3.2-nemoretriever-1b-vlm-embed-v1",
			"nvidia/llama-3.2-nv-embedqa-1b-v1",
			"nvidia/llama-3.3-nemotron-super-49b-v1",
			"nvidia/llama-3.3-nemotron-super-49b-v1.5",
			"nvidia/llama-nemotron-embed-1b-v2",
			"nvidia/llama-nemotron-embed-vl-1b-v2",
			"nvidia/llama3-chatqa-1.5-70b",
			"nvidia/mistral-nemo-minitron-8b-8k-instruct",
			"nvidia/nemoretriever-parse",
			"nvidia/nemotron-3-content-safety",
			"nvidia/nemotron-3-nano-30b-a3b",
			"nvidia/nemotron-3-nano-omni-30b-a3b-reasoning",
			"nvidia/nemotron-3-super-120b-a12b",
			"nvidia/nemotron-3-ultra-550b-a55b",
			"nvidia/nemotron-3.5-content-safety",
			"nvidia/nemotron-4-340b-instruct",
			"nvidia/nemotron-4-340b-reward",
			"nvidia/nemotron-content-safety-reasoning-4b",
			"nvidia/nemotron-mini-4b-instruct",
			"nvidia/nemotron-nano-12b-v2-vl",
			"nvidia/nemotron-nano-3-30b-a3b",
			"nvidia/nemotron-parse",
			"nvidia/neva-22b",
			"nvidia/nv-embed-v1",
			"nvidia/nv-embedcode-7b-v1",
			"nvidia/nv-embedqa-e5-v5",
			"nvidia/nv-embedqa-mistral-7b-v2",
			"nvidia/nvclip",
			"nvidia/nvidia-nemotron-nano-9b-v2",
			"nvidia/riva-translate-4b-instruct",
			"nvidia/riva-translate-4b-instruct-v1.1",
			"nvidia/vila",
			// OpenAI
			"openai/gpt-oss-120b",
			"openai/gpt-oss-20b",
			// Qwen
			"qwen/qwen3-next-80b-a3b-instruct",
			"qwen/qwen3.5-122b-a10b",
			"qwen/qwen3.5-397b-a17b",
			// Sarvam AI
			"sarvamai/sarvam-m",
			// Snowflake
			"snowflake/arctic-embed-l",
			// StepFun AI
			"stepfun-ai/step-3.5-flash",
			"stepfun-ai/step-3.7-flash",
			// Stockmark
			"stockmark/stockmark-2-100b-instruct",
			// Upstage
			"upstage/solar-10.7b-instruct",
			// Writer
			"writer/palmyra-creative-122b",
			"writer/palmyra-fin-70b-32k",
			"writer/palmyra-med-70b",
			"writer/palmyra-med-70b-32k",
			// Z.AI
			"z-ai/glm-5.1",
			// Zyphra
			"zyphra/zamba2-7b-instruct",
		},
	}

	if cfg.NvidiaAPIKey == "" {
		log.Println("WARNING: NVIDIA_API_KEY is not set. Set it in backend/.env")
	}

	return cfg
}
