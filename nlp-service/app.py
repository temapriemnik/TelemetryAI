import os
from flask import Flask, request, jsonify
import joblib
import numpy as np

app = Flask(__name__)

MODEL_PATH = os.environ.get('MODEL_PATH', './smart_log_model.pkl')

vectorizer, model = joblib.load(MODEL_PATH)

@app.route('/health', methods=['GET'])
def health():
    return jsonify({'status': 'ok'})

@app.route('/predict', methods=['POST'])
def predict():
    data = request.get_json()
    if not data or 'log' not in data:
        return jsonify({'error': 'missing "log" field'}), 400
    
    log_text = data['log']
    X = vectorizer.transform([log_text])
    label = model.predict(X)[0]
    
    return jsonify({'label': label})

if __name__ == '__main__':
    app.run(host='0.0.0.0', port=int(os.environ.get('PORT', 5000)))