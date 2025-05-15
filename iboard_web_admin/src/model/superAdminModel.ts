export class SuperAdmin {
    id: number | null;
    email: string | null;
    password: string | null;

    constructor({ id, email, password }: SuperAdmin) {
        this.id = id;
        this.email = email;
        this.password = password;
    }

    static validate(admin: SuperAdmin): boolean {
        const emailPattern = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;

        if (!admin.email || !emailPattern.test(admin.email)) {
            return false;
        }

        if (!admin.password) {
            return false;
        }

        return true;
    }
}

