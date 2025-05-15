import { File } from './fileModel';
import type { Dayjs } from 'dayjs';

export enum NoticeTypes {
    Government = 'government',
    Urgent = 'urgent',
    Normal = 'normal',
    Building = 'building'
}

export enum NoticeStatus {
    Active = 'active',
    Inactive = 'inactive'
}

export class Notice {
    id: number | null | undefined;
    title: string | null | undefined;
    description: string | null | undefined;
    type: NoticeTypes | null | undefined;
    status: NoticeStatus | null | undefined;
    startTime: Dayjs | null | undefined;
    endTime: Dayjs | null | undefined;
    file: File | null | undefined;
    isPublic: boolean | undefined;

    constructor({ id, title, description, type, status, startTime, endTime, file, isPublic }: Partial<Notice> = {}) {
        this.id = id;
        this.title = title;
        this.description = description;
        this.type = type;
        this.status = status;
        this.startTime = startTime;
        this.endTime = endTime;
        this.file = file;
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

        if (!this.status) {
            return false;
        }

        if (!this.startTime) {
            return false;
        }

        if (!this.endTime) {
            return false;
        }

        if (this.isPublic === undefined) {
            return false;
        }

        return true;
    }
}
