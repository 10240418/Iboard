export class Api {
    id: number | null;
    name: string | null;
    endpoint: string | null;
    method: string | null;
    description: string | null;

    constructor({ id, name, endpoint, method, description }: Api) {
        this.id = id;
        this.name = name;
        this.endpoint = endpoint;
        this.method = method;
        this.description = description;
    }

    static validate(api: Api): boolean {
        if (!api.id) {
            return false;
        }

        if (!api.name) {
            return false;
        }

        if (!api.endpoint) {
            return false;
        }

        if (!api.method) {
            return false;
        }

        if (!api.description) {
            return false;
        }

        return true;
    }
}
