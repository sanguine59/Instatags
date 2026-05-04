FROM python:3.10-slim

RUN apt-get update && apt-get install -y --no-install-recommends \
    build-essential \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

ENV HF_HOME=/app/.cache/huggingface

COPY services/ai-service/requirements.txt ./requirements.txt
RUN --mount=type=cache,target=/root/.cache/pip \
    pip install -r requirements.txt

RUN python -c "from transformers import CLIPModel, CLIPProcessor; \
    CLIPProcessor.from_pretrained('openai/clip-vit-base-patch32'); \
    CLIPModel.from_pretrained('openai/clip-vit-base-patch32')"

COPY protos/ ./protos/

COPY services/ai-service/ ./services/ai-service/

WORKDIR /app/services/ai-service

RUN python -m grpc_tools.protoc \
       -I../../protos/ai \
       --python_out=. \
       --grpc_python_out=. \
       ../../protos/ai/ai.proto

CMD ["python", "main.py"]