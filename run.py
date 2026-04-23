import sys
sys.path.insert(0, '/home/artem/Desktop/TelemetryAI')

import uvicorn
from backend.app.main import app

if __name__ == "__main__":
    uvicorn.run(app, host="0.0.0.0", port=8000)