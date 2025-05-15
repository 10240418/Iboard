import { Advertisement } from "./advertisementModel";
import { Device } from "./deviceModel";
import { Notice } from "./noticeModel";


export class Building {
    id: number | null | undefined;
    name: string | null | undefined;
    ismartId: string | null | undefined;
    location: string | null | undefined;
    remark: string | null | undefined;
    notices?: Notice[];
    advertisements?: Advertisement[];
    devices?: Device[];

    constructor({ id, name, ismartId, location, remark, notices, advertisements, devices }: Building) {
        this.id = id;
        this.name = name;
        this.ismartId = ismartId;
        this.location = location;
        this.remark = remark;
        this.notices = notices;
        this.advertisements = advertisements;
        this.devices = devices;
    }

    static validate(building: Building): boolean {
        if (!building.name) {
            return false;
        }

        if (!building.ismartId) {
            return false;
        }

        if (!building.location) {
            return false;
        }

        if (!building.remark) {
            return false;
        }

        return true;
    }
}
