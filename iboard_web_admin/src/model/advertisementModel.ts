import { File as FileModel } from './fileModel';
import type { Dayjs } from 'dayjs';

export enum AdvertisementTypes {
    Video = 'video',
    Image = 'image'
}

export enum DisplayTypes {
    Top = 'top',
    Full = 'full',
    TopFull = 'topfull'
}

export enum AdvertisementStatus {
    Active = 'active',
    Inactive = 'inactive'
}

export class Advertisement {
    id: number | undefined;
    title: string | undefined;
    description: string | undefined;
    type: AdvertisementTypes | undefined;
    file: FileModel | null | undefined;
    duration: number | undefined | null;
    startTime: Dayjs | undefined | null;
    endTime: Dayjs | undefined | null;
    status: AdvertisementStatus | undefined | null;
    display: DisplayTypes | undefined;
    isPublic: boolean | undefined;

    constructor({ id, title, description, type, file, duration, startTime, endTime, status, display, isPublic }: Partial<Advertisement> = {}) {
        this.id = id;
        this.title = title;
        this.description = description;
        this.type = type;
        this.file = file;
        this.duration = duration;
        this.startTime = startTime;
        this.endTime = endTime;
        this.status = status;
        this.display = display;
        this.isPublic = isPublic;
    }

    public validate(): boolean {
        if (!this.title) {
            return false;
        }

        if (!this.description) {
            return false;
        }

        if (!this.type) {
            return false;
        }

        // const urlPattern = /^(http|https):\/\/[^\s]+$/;
        if (!this.file) {
            return false;
        } else if (this.file instanceof FileModel) {
            return true
        }

        if (!this.duration) {
            return false;
        }

        if (!this.startTime || !(this.startTime instanceof Date)) {
            return false;
        }

        if (!this.endTime || !(this.endTime instanceof Date)) {
            return false;
        }

        if (!this.status) {
            return false;
        }

        if (!this.display) {
            return false;
        }

        if (this.isPublic === undefined) {
            return false;
        }

        return true;
    }
}
