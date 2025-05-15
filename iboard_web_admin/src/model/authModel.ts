export class LoginModel {
    email: string;
    password: string;

    constructor({ email, password }: { email: string; password: string }) {
        this.email = email;
        this.password = password;
    }

    static validate(auth: LoginModel): boolean {
        if (!auth.email) {
            return false;
        }

        if (!auth.password) {
            return false;
        }

        return true;
    }
}
