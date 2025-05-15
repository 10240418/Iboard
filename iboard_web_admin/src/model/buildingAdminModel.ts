export enum BuildingAdminStatus {
    Active = "active",
    Inactive = "inactive",
}

export class BuildingAdmin {
    id: number | null;
    email: string | null;
    password: string | null;
    status: string | null;

    constructor({ id, email, password, status }: BuildingAdmin) {
        this.id = id;
        this.email = email;
        this.password = password;
        this.status = status;
    }

    static validate(admin: BuildingAdmin): boolean {
        const emailPattern = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;

        if (!admin.email || !emailPattern.test(admin.email)) {
            return false;
        }

        if (!admin.password) {
            return false;
        }

        if (!admin.status) {
            return false;
        }

        return true;
    }
}
