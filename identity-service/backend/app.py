from flask import Flask, request, jsonify, send_from_directory
from flask_jwt_extended import (
    JWTManager, create_access_token, create_refresh_token,
    jwt_required, get_jwt_identity
)
from flask_cors import CORS
from models import db, User, Project, ApiKey
from datetime import timedelta
from functools import wraps
import os

BASE_DIR = os.path.dirname(os.path.abspath(__file__))
FRONTEND_DIR = os.path.join(os.path.dirname(BASE_DIR), 'frontend')

app = Flask(__name__, static_folder=FRONTEND_DIR, static_url_path='')

CORS(app, origins='*')

# Конфигурация
app.config['SQLALCHEMY_DATABASE_URI'] = 'sqlite:///app.db'
app.config['SQLALCHEMY_TRACK_MODIFICATIONS'] = False
app.config['JWT_SECRET_KEY'] = 'super-secret-change-this-in-production'
app.config['JWT_ACCESS_TOKEN_EXPIRES'] = timedelta(hours=1)
app.config['JWT_REFRESH_TOKEN_EXPIRES'] = timedelta(days=30)

jwt = JWTManager(app)
db.init_app(app)

with app.app_context():
    db.create_all()

def require_api_key(f):
    @wraps(f)
    def decorated(*args, **kwargs):
        key = request.headers.get('X-API-Key')
        if not key:
            return jsonify({'error': 'API key required'}), 401
        api_key = ApiKey.query.filter_by(key=key, is_active=True).first()
        if not api_key:
            return jsonify({'error': 'Invalid or inactive API key'}), 403
        request.user_id = api_key.user_id
        return f(*args, **kwargs)
    return decorated

@app.route('/')
def index():
    return send_from_directory(FRONTEND_DIR, 'index.html')

@app.route('/<path:filename>')
def serve_static(filename):
    return send_from_directory(FRONTEND_DIR, filename)

@app.route('/api/auth/register', methods=['POST'])
def register():
    data = request.get_json()
    if not data:
        return jsonify({'error': 'Нет данных'}), 400

    required = ['email', 'password', 'name']
    if not all(k in data for k in required):
        return jsonify({'error': f'Отсутствуют поля: {required}'}), 400

    if User.query.filter_by(email=data['email']).first():
        return jsonify({'error': 'Email уже зарегистрирован'}), 409

    user = User(email=data['email'], name=data['name'])
    user.password = data['password']
    db.session.add(user)
    db.session.commit()
    return jsonify(user.to_dict()), 201

@app.route('/api/auth/login', methods=['POST'])
def login():
    data = request.get_json()
    user = User.query.filter_by(email=data.get('email')).first()
    if not user or not user.check_password(data.get('password', '')):
        return jsonify({'error': 'Неверные учетные данные'}), 401

    access_token = create_access_token(identity=str(user.id))
    refresh_token = create_refresh_token(identity=str(user.id))
    return jsonify({
        'access_token': access_token,
        'refresh_token': refresh_token,
        'user': user.to_dict()
    })

@app.route('/api/auth/refresh', methods=['POST'])
@jwt_required(refresh=True)
def refresh():
    identity = get_jwt_identity()
    access_token = create_access_token(identity=identity)
    return jsonify({'access_token': access_token})

@app.route('/api/profile', methods=['GET'])
@jwt_required()
def get_profile():
    user_id = get_jwt_identity()
    user = User.query.get(user_id)
    return jsonify(user.to_dict())

@app.route('/api/profile', methods=['PUT'])
@jwt_required()
def update_profile():
    user_id = get_jwt_identity()
    user = User.query.get(user_id)
    data = request.get_json()
    if 'name' in data:
        user.name = data['name']
    if 'email' in data:
        existing = User.query.filter(User.email == data['email'], User.id != user_id).first()
        if existing:
            return jsonify({'error': 'Email уже используется'}), 409
        user.email = data['email']
    if 'password' in data:
        user.password = data['password']
    db.session.commit()
    return jsonify(user.to_dict())

@app.route('/api/projects', methods=['POST'])
@jwt_required()
def create_project():
    user_id = get_jwt_identity()
    data = request.get_json()
    if 'name' not in data:
        return jsonify({'error': 'Не указано название проекта'}), 400
    project = Project(name=data['name'], user_id=user_id)
    db.session.add(project)
    db.session.commit()
    return jsonify(project.to_dict()), 201

@app.route('/api/projects', methods=['GET'])
@jwt_required()
def list_projects():
    user_id = get_jwt_identity()
    projects = Project.query.filter_by(user_id=user_id).all()
    return jsonify([p.to_dict() for p in projects])

@app.route('/api/projects/<int:project_id>', methods=['GET'])
@jwt_required()
def get_project(project_id):
    user_id = get_jwt_identity()
    project = Project.query.filter_by(id=project_id, user_id=user_id).first()
    if not project:
        return jsonify({'error': 'Проект не найден'}), 404
    return jsonify(project.to_dict())

@app.route('/api/projects/<int:project_id>', methods=['DELETE'])
@jwt_required()
def delete_project(project_id):
    user_id = get_jwt_identity()
    project = Project.query.filter_by(id=project_id, user_id=user_id).first()
    if not project:
        return jsonify({'error': 'Проект не найден'}), 404
    db.session.delete(project)
    db.session.commit()
    return jsonify({'message': 'Проект удалён'})

@app.route('/api/api-keys', methods=['POST'])
@jwt_required()
def create_api_key():
    user_id = get_jwt_identity()
    data = request.get_json()
    name = data.get('name', 'Безымянный ключ')
    api_key = ApiKey(name=name, user_id=user_id)
    db.session.add(api_key)
    db.session.commit()
    return jsonify(api_key.to_dict()), 201

@app.route('/api/api-keys', methods=['GET'])
@jwt_required()
def list_api_keys():
    user_id = get_jwt_identity()
    keys = ApiKey.query.filter_by(user_id=user_id).all()
    return jsonify([k.to_dict() for k in keys])

@app.route('/api/api-keys/<int:key_id>', methods=['DELETE'])
@jwt_required()
def revoke_api_key(key_id):
    user_id = get_jwt_identity()
    key = ApiKey.query.filter_by(id=key_id, user_id=user_id).first()
    if not key:
        return jsonify({'error': 'Ключ не найден'}), 404
    key.is_active = False
    db.session.commit()
    return jsonify({'message': 'Ключ отозван'})

@app.route('/api/agent/data', methods=['GET'])
@require_api_key
def agent_data():
    user_id = request.user_id
    projects = Project.query.filter_by(user_id=user_id).all()
    return jsonify({
        'user_id': user_id,
        'projects': [p.to_dict() for p in projects]
    })

@app.route('/api/agent/projects', methods=['POST'])
@require_api_key
def agent_create_project():
    user_id = request.user_id
    data = request.get_json()
    if 'name' not in data:
        return jsonify({'error': 'Не указано название проекта'}), 400
    project = Project(name=data['name'], user_id=user_id)
    db.session.add(project)
    db.session.commit()
    return jsonify(project.to_dict()), 201

if __name__ == '__main__':
    app.run(debug=True)