import { SuperAdmin } from './superAdminModel';
import { BuildingAdmin } from './buildingAdminModel';

export class File {
    id: number | undefined;
    md5: string | null | undefined;
    mimeType: string | undefined;
    size: number | undefined;
    path: string | undefined;
    ossRegion: string | null | undefined;
    ossBucket: string | null | undefined;
    uploader: SuperAdmin | BuildingAdmin | undefined;
    uploaderType: string | undefined;

    constructor({ id, md5, mimeType, size, path, ossRegion, ossBucket, uploader, uploaderType }: Partial<File> = {}) {
        this.id = id;
        this.md5 = md5;
        this.mimeType = mimeType;
        this.size = size;
        this.path = path;
        this.ossRegion = ossRegion;
        this.ossBucket = ossBucket;
        this.uploader = uploader;
        this.uploaderType = uploaderType;
    }

    public validate(): boolean {
        if (!this.md5) {
            return false;
        }

        if (!this.mimeType) {
            return false;
        }

        if (!this.size) {
            return false;
        }

        if (!this.path) {
            return false;
        }

        if (!this.ossRegion) {
            return false;
        }

        if (!this.ossBucket) {
            return false;
        }

        if (!this.uploader || !(this.uploader instanceof SuperAdmin || this.uploader instanceof BuildingAdmin)) {
            return false;
        }

        if (!this.uploaderType) {
            return false;
        }

        return true;
    }
}

export class AliyunOSSPolicy {
    accessId: string | null;
    host: string | null;
    policy: string | null;
    signature: string | null;
    expire: string | null;
    dir: string | null;

    constructor({ accessId, host, policy, signature, expire, dir }: AliyunOSSPolicy) {
        this.accessId = accessId;
        this.host = host;
        this.policy = policy;
        this.signature = signature;
        this.expire = expire;
        this.dir = dir;
    }

    static validate(policy: AliyunOSSPolicy): boolean {
        if (!policy.accessId) {
            return false;
        }

        if (!policy.host) {
            return false;
        }

        if (!policy.policy) {
            return false;
        }

        if (!policy.signature) {
            return false;
        }

        if (!policy.expire) {
            return false;
        }

        if (!policy.dir) {
            return false;
        }

        return true;
    }
}