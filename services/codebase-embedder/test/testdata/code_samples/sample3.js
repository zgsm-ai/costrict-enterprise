class UserService {
    constructor() {
        this.users = new Map();
    }
    
    addUser(id, name) {
        this.users.set(id, name);
    }
    
    getUser(id) {
        return this.users.get(id);
    }
    
    listUsers() {
        return Object.fromEntries(this.users);
    }
    
    deleteUser(id) {
        return this.users.delete(id);
    }
    
    hasUser(id) {
        return this.users.has(id);
    }
    
    getUserCount() {
        return this.users.size;
    }
}

function main() {
    const service = new UserService();
    service.addUser("1", "Alice");
    service.addUser("2", "Bob");
    
    const user = service.getUser("1");
    if (user) {
        console.log(`User found: ${user}`);
    }
    
    const users = service.listUsers();
    for (const [id, name] of Object.entries(users)) {
        console.log(`ID: ${id}, Name: ${name}`);
    }
    
    console.log(`User count: ${service.getUserCount()}`);
    console.log(`Has user 2: ${service.hasUser("2")}`);
    
    service.deleteUser("2");
    console.log(`After deletion: ${service.listUsers()}`);
}

main();