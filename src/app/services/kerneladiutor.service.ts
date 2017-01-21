import { Injectable } from '@angular/core'
import { Http, Response, URLSearchParams } from '@angular/http'

import { Observable } from 'rxjs/Observable'

export class KernelAdiutorDevice {

    constructor(private device: any) { }

    getId(): string {
        return this.device.id
    }

    getAndroidVersion(): string {
        return this.device.android_version
    }

    getKernelVersion(): string {
        return this.device.kernel_version
    }

    getAppVersion(): string {
        return this.device.app_version
    }

    getBoard(): string {
        return this.device.board
    }

    getModel(): string {
        return this.device.model
    }

    getVendor(): string {
        let vendor = <string>this.device.vendor
        return vendor.charAt(0).toUpperCase() + vendor.substr(1)
    }

    getCpuInfo(): string {
        return this.device.cpuinfo
    }

    getCommands(): string[] {
        return this.device.commands
    }

    getTimes(): number[] {
        return this.device.times
    }

    getCpu(): number {
        return this.device.cpu
    }

    getScore(): number {
        return this.device.score
    }

}

@Injectable()
export class KernelAdiutorService {

    getDevicesLink: string = "/kerneladiutor/api/v1/device/get?"

    constructor(
        private http: Http
    ) { }

    getDevices(page?: number): Observable<KernelAdiutorDevice> {
        let url = new URLSearchParams(this.getDevicesLink)
        if (page > 0) {
            url.set('page', page.toString())
        }

        let observer = new Observable<KernelAdiutorDevice>((observer: any) => {
            this.http.get(url.toString()).forEach((response: Response) => {
                if (response) {
                    let devices = response.json()
                    for (let device of devices) {
                        observer.next(new KernelAdiutorDevice(device))
                    }
                } else {
                    observer.next(null)
                }
            })
        })

        return observer
    }

}
