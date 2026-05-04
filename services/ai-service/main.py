import grpc
from concurrent import futures
import io

import torch
import torch.nn.functional as F
from transformers import CLIPModel, CLIPProcessor
from PIL import Image

import ai_pb2 as service_pb2
import ai_pb2_grpc as service_pb2_grpc

INDEX_PATH = "./.model/clip_index.pt"
CLIP_MODEL_ID = "openai/clip-vit-base-patch32"
DEVICE = torch.device("cpu")

class AIService(service_pb2_grpc.AIServiceServicer):
    def __init__(self):
        print("Initializing AI Service...")
        self.model = self.load_model()
        self.processor = CLIPProcessor.from_pretrained(CLIP_MODEL_ID, local_files_only=True)
        self.text_embeddings, self.tags = self.load_index()
        print("AI Service Ready.")

    def load_model(self):
        try:
            model = CLIPModel.from_pretrained(CLIP_MODEL_ID, local_files_only=True).to(DEVICE)
            model.eval()
            return model
        except Exception as e:
            print(f"CRITICAL ERROR: Failed to load CLIP model: {e}")
            return None

    def load_index(self):
        try:
            data = torch.load(INDEX_PATH, map_location=DEVICE)
            embeddings = data["embeddings"].to(DEVICE)
            tags = data["tags"]
            embeddings = F.normalize(embeddings, dim=-1)
            print("Loaded index:", embeddings.shape)
            return embeddings, tags
        except FileNotFoundError:
            print("CRITICAL ERROR: clip_index.pt not found!")
            return None, []
        except Exception as e:
            print(f"CRITICAL ERROR: Failed to load index: {e}")
            return None, []

    def PredictHashtags(self, request, context):
        if self.model is None or self.text_embeddings is None:
            context.set_code(grpc.StatusCode.INTERNAL)
            context.set_details('Model or index not loaded correctly')
            return service_pb2.PredictResponse()

        try:
            image = Image.open(io.BytesIO(request.image_data)).convert("RGB")
            inputs = self.processor(images=image, return_tensors="pt").to(DEVICE)

            with torch.no_grad():
                image_features = self.model.get_image_features(**inputs)
            image_features = F.normalize(image_features, dim=-1)

            sims = (image_features @ self.text_embeddings.T).squeeze(0)
            top_k = min(10, sims.shape[0])
            top_indices = sims.topk(top_k).indices

            recommended = [self.tags[idx] for idx in top_indices.tolist()]
            return service_pb2.PredictResponse(hashtags=recommended)

        except Exception as e:
            print(f"Prediction Error: {e}")
            context.set_code(grpc.StatusCode.INTERNAL)
            context.set_details(str(e))
            return service_pb2.PredictResponse()

def serve():
    server = grpc.server(futures.ThreadPoolExecutor(max_workers=4))
    service_pb2_grpc.add_AIServiceServicer_to_server(AIService(), server)
    
    server.add_insecure_port('[::]:50051')
    print("AI Service listening on port 50051...")
    server.start()
    server.wait_for_termination()

if __name__ == '__main__':
    serve()