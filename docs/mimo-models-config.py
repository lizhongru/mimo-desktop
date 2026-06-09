# MiMo Model Configuration
# Source: https://platform.xiaomimimo.com/docs/zh-CN/quick-start/model
# Updated: 2026-06-09

MIMO_MODELS = {
    # Text Generation Models - Pro Series
    "mimo-v2.5-pro": {
        "series": "Pro",
        "capabilities": ["text_generation", "deep_thinking", "streaming", "function_calling", "structured_output", "web_search"],
        "context_window": 1000000,  # 1M tokens
        "max_output": 128000,  # 128K tokens
        "max_rpm": 100,
        "max_tpm": 10000000,  # 10M tokens
        "description": "复杂推理、深度分析、长文档处理",
    },
    "mimo-v2-pro": {
        "series": "Pro",
        "capabilities": ["text_generation", "deep_thinking", "streaming", "function_calling", "structured_output", "web_search"],
        "context_window": 1000000,
        "max_output": 128000,
        "max_rpm": 100,
        "max_tpm": 10000000,
        "description": "复杂推理、深度分析、长文档处理",
    },
    
    # Text Generation Models - Omni Series
    "mimo-v2.5": {
        "series": "Omni",
        "capabilities": ["text_generation", "multimodal", "deep_thinking", "streaming", "function_calling", "structured_output", "web_search"],
        "context_window": 1000000,
        "max_output": 128000,
        "max_rpm": 100,
        "max_tpm": 10000000,
        "description": "图片、音频、视频内容理解",
    },
    "mimo-v2-omni": {
        "series": "Omni",
        "capabilities": ["text_generation", "multimodal", "deep_thinking", "streaming", "function_calling", "structured_output", "web_search"],
        "context_window": 256000,
        "max_output": 128000,
        "max_rpm": 100,
        "max_tpm": 10000000,
        "description": "图片、音频、视频内容理解",
    },
    
    # Text Generation Models - Flash Series
    "mimo-v2-flash": {
        "series": "Flash",
        "capabilities": ["text_generation", "deep_thinking", "streaming", "function_calling", "structured_output", "web_search"],
        "context_window": 256000,
        "max_output": 64000,
        "max_rpm": 100,
        "max_tpm": 10000000,
        "description": "低成本、快速响应",
    },
    
    # ASR Models
    "mimo-v2.5-asr": {
        "series": "ASR",
        "capabilities": ["speech_recognition"],
        "context_window": 8000,
        "max_output": 2000,
        "max_rpm": 100,
        "max_tpm": 10000,
        "description": "语音转文字（支持中英双语）",
    },
    
    # TTS Models
    "mimo-v2.5-tts": {
        "series": "TTS",
        "capabilities": ["speech_synthesis"],
        "context_window": 8000,
        "max_output": 8000,
        "max_rpm": 100,
        "max_tpm": 10000000,
        "description": "文字转语音（标准预置音色）",
    },
    "mimo-v2.5-tts-voiceclone": {
        "series": "TTS",
        "capabilities": ["speech_synthesis", "voice_clone"],
        "context_window": 8000,
        "max_output": 8000,
        "max_rpm": 100,
        "max_tpm": 10000000,
        "description": "声音克隆（上传音频样本）",
    },
    "mimo-v2.5-tts-voicedesign": {
        "series": "TTS",
        "capabilities": ["speech_synthesis", "voice_design"],
        "context_window": 8000,
        "max_output": 8000,
        "max_rpm": 100,
        "max_tpm": 10000000,
        "description": "自定义音色设计",
    },
    "mimo-v2-tts": {
        "series": "TTS",
        "capabilities": ["speech_synthesis"],
        "context_window": 8000,
        "max_output": 8000,
        "max_rpm": 100,
        "max_tpm": 10000000,
        "description": "文字转语音",
    },
}

# Quick Selection Guide
RECOMMENDED_MODELS = {
    "complex_reasoning": "mimo-v2.5-pro",
    "multimodal": "mimo-v2.5",
    "fast_response": "mimo-v2-flash",
    "speech_to_text": "mimo-v2.5-asr",
    "text_to_speech": "mimo-v2.5-tts",
    "voice_clone": "mimo-v2.5-tts-voiceclone",
    "voice_design": "mimo-v2.5-tts-voicedesign",
}

# Model series groups for UI display
MODEL_SERIES = {
    "Pro": {
        "name": "Pro 系列",
        "description": "复杂推理、深度分析",
        "models": ["mimo-v2.5-pro", "mimo-v2-pro"],
    },
    "Omni": {
        "name": "Omni 系列",
        "description": "全模态理解",
        "models": ["mimo-v2.5", "mimo-v2-omni"],
    },
    "Flash": {
        "name": "Flash 系列",
        "description": "快速响应、低成本",
        "models": ["mimo-v2-flash"],
    },
    "ASR": {
        "name": "语音识别",
        "description": "语音转文字",
        "models": ["mimo-v2.5-asr"],
    },
    "TTS": {
        "name": "语音合成",
        "description": "文字转语音",
        "models": ["mimo-v2.5-tts", "mimo-v2.5-tts-voiceclone", "mimo-v2.5-tts-voicedesign", "mimo-v2-tts"],
    },
}
