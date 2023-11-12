import requests
import threading
import random
import time

SERVER_URL = "http://localhost:8080"

def create_game():
    response = requests.post(f"{SERVER_URL}/createGame")
    return response.json()['gameCode']

def join_game(game_code):
    response = requests.post(f"{SERVER_URL}/joinGame/{game_code}")
    return response.json()['userId']

def complete_game(game_code, user_id, score):
    response = requests.post(f"{SERVER_URL}/completeGame/{game_code}?user_id=user{user_id}&score={score}")
    return response.json()

def simulate_player(game_code, player_name, score):
    time.sleep(random.randint(0, 3))
    print(f"{player_name} joining...")
    user_id = join_game(game_code)
    print(f"{player_name} joined as user{user_id}")

    time.sleep(random.randint(4, 6))
    print(f"{player_name}'s score: {score}")
    result = complete_game(game_code, user_id, score)
    print(f"{player_name}'s result: {result}")

def test_game_flow():
    game_code = create_game()
    print(f"Game created with code: {game_code}")

    # Simulate two players joining and completing the game
    player_a = threading.Thread(target=simulate_player, args=(game_code, "Alice", random.randint(0, 100)))
    player_b = threading.Thread(target=simulate_player, args=(game_code, "Bob", random.randint(0, 100)))

    player_a.start()
    player_b.start()

    player_a.join()
    player_b.join()

if __name__ == "__main__":
    test_game_flow()
