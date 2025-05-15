export class DeviceSettings {
    arrearageUpdateDuration: number | undefined;
    noticeUpdateDuration: number | undefined;
    advertisementUpdateDuration: number | undefined;
    advertisementPlayDuration: number | undefined;
    noticePlayDuration: number | undefined;
    spareDuration: number | undefined;
    noticeStayDuration: number | undefined;

    constructor({
        arrearageUpdateDuration,
        noticeUpdateDuration,
        advertisementUpdateDuration,
        advertisementPlayDuration,
        noticePlayDuration,
        spareDuration,
        noticeStayDuration
    }: Partial<DeviceSettings> = {}) {
        this.arrearageUpdateDuration = arrearageUpdateDuration;
        this.noticeUpdateDuration = noticeUpdateDuration;
        this.advertisementUpdateDuration = advertisementUpdateDuration;
        this.advertisementPlayDuration = advertisementPlayDuration;
        this.noticePlayDuration = noticePlayDuration;
        this.spareDuration = spareDuration;
        this.noticeStayDuration = noticeStayDuration;
    }

    public validate(): boolean {
        return typeof this.arrearageUpdateDuration === 'number' &&
            typeof this.noticeUpdateDuration === 'number' &&
            typeof this.advertisementUpdateDuration === 'number' &&
            typeof this.advertisementPlayDuration === 'number' &&
            typeof this.noticePlayDuration === 'number' &&
            typeof this.spareDuration === 'number' &&
            typeof this.noticeStayDuration === 'number';
    }
}

export enum DeviceStatus {
    Online = 'online',
    Offline = 'offline',
    Unknown = 'unknown'
}

export class Device {
    id: number | undefined;
    deviceId: string | undefined;
    buildingId: number | undefined;
    settings: DeviceSettings | undefined;
    buildingName: string | undefined;
    status: DeviceStatus | undefined;

    constructor({ id, deviceId, buildingId, settings, buildingName, status }: Partial<Device> = {}) {
        this.id = id;
        this.deviceId = deviceId;
        this.buildingId = buildingId;
        this.settings = new DeviceSettings(settings);
        this.buildingName = buildingName;
        this.status = status;
    }

    public validate(): boolean {
        if (!this.deviceId) {
            return false;
        }

        if (!this.buildingId) {
            return false;
        }

        if (!this.settings) {
            return false;
        }

        if (!this.status) {
            return false;
        }

        return true;
    }
}
