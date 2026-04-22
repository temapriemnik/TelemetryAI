import requests
import os
import json
import sys

BASE_URL = "http://127.0.0.1:5000/api"
CONFIG_DIR = os.path.expanduser("~/.myagent")
CONFIG_FILE = os.path.join(CONFIG_DIR, "config.json")

def load_config():
    if os.path.exists(CONFIG_FILE):
        with open(CONFIG_FILE, 'r') as f:
            return json.load(f)
    return {}

def save_config(api_key):
    os.makedirs(CONFIG_DIR, exist_ok=True)
    with open(CONFIG_FILE, 'w') as f:
        json.dump({"api_key": api_key}, f)

def clear_config():
    if os.path.exists(CONFIG_FILE):
        os.remove(CONFIG_FILE)

def get_projects(api_key):
    headers = {"X-API-Key": api_key}
    resp = requests.get(f"{BASE_URL}/agent/data", headers=headers)
    resp.raise_for_status()
    return resp.json()

def create_project(api_key, name):
    headers = {"X-API-Key": api_key, "Content-Type": "application/json"}
    resp = requests.post(f"{BASE_URL}/agent/projects", headers=headers, json={"name": name})
    resp.raise_for_status()
    return resp.json()

def main():
    config = load_config()
    api_key = config.get("api_key")

    if not api_key:
        print("API-ключ не найден.")
        api_key = input("Введите ваш API-ключ: ").strip()
        save_config(api_key)
        print("Ключ сохранён.")

    # Проверяем соединение
    try:
        data = get_projects(api_key)
        print(f"✅ Подключено! User ID: {data['user_id']}")
        if data['projects']:
            print("Ваши проекты:")
            for p in data['projects']:
                print(f"  - {p['name']} (ID: {p['id']})")
        else:
            print("У вас пока нет проектов.")
    except requests.exceptions.HTTPError as e:
        if e.response.status_code in (401, 403):
            print("❌ Неверный или неактивный API-ключ.")
            clear_config()
            sys.exit(1)
        else:
            print(f"Ошибка: {e}")
            sys.exit(1)
    except Exception as e:
        print(f"Ошибка соединения: {e}")
        sys.exit(1)

    # Меню
    while True:
        print("\nКоманды: [list] показать проекты, [create <название>] создать проект, [quit] выход")
        cmd = input("> ").strip().split(maxsplit=1)
        if not cmd:
            continue
        action = cmd[0].lower()
        if action == "quit" or action == "exit":
            break
        elif action == "list":
            try:
                data = get_projects(api_key)
                if data['projects']:
                    for p in data['projects']:
                        print(f"  - {p['name']} (ID: {p['id']})")
                else:
                    print("Проектов нет.")
            except Exception as e:
                print(f"Ошибка: {e}")
        elif action == "create":
            if len(cmd) < 2:
                print("Укажите название проекта: create <название>")
                continue
            name = cmd[1]
            try:
                p = create_project(api_key, name)
                print(f"✅ Проект '{p['name']}' создан (ID: {p['id']})")
            except Exception as e:
                print(f"Ошибка создания: {e}")
        else:
            print("Неизвестная команда")

if __name__ == "__main__":
    main()