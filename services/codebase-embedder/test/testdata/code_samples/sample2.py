class UserService:
    def __init__(self):
        self.users = {}
    
    def add_user(self, user_id, name):
        self.users[user_id] = name
    
    def get_user(self, user_id):
        return self.users.get(user_id)
    
    def list_users(self):
        return self.users
    
    def delete_user(self, user_id):
        if user_id in self.users:
            del self.users[user_id]
            return True
        return False

def main():
    service = UserService()
    service.add_user("1", "Alice")
    service.add_user("2", "Bob")
    
    user = service.get_user("1")
    if user:
        print(f"User found: {user}")
    
    users = service.list_users()
    for user_id, name in users.items():
        print(f"ID: {user_id}, Name: {name}")
    
    service.delete_user("2")
    print(f"After deletion: {service.list_users()}")

if __name__ == "__main__":
    main()