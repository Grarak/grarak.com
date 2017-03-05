import { Component } from '@angular/core'
import { ActivatedRoute, Params, Router } from '@angular/router'

import { KernelAdiutorService, KernelAdiutorDevice } from '../../services/kerneladiutor.service'

@Component({
    selector: `kerneladiutor-deviceinfo-page`,
    template: `
        <pageparent-view [parentMarginOffset]="0.05">
            <card-view>
                <card-title>
                    {{device?.getVendor()}} {{device?.getModel()}}
                </card-title>
                <card-content>
                    <div *ngFor="let content of contentList; let i = index">
                        <span [innerHTML]="newLine(i)"></span><strong>{{content[0]}}</strong><br><span style="word-wrap:break-word" [innerHTML]="content[1]"></span>
                    </div>
                </card-content>
            </card-view>
        </pageparent-view>
    `
})
export class KernelAdiutorDeviceInfoComponent {

    device: KernelAdiutorDevice
    contentList: string[][]

    constructor(
        private kaService: KernelAdiutorService,
        private router: Router,
        private route: ActivatedRoute
    ) { }

    ngOnInit() {
        this.route.params.forEach((params: Params) => {
            this.kaService.getDeviceById(params['id']).forEach((device: KernelAdiutorDevice) => {
                if (device == null) {
                    this.router.navigate(['404'])
                } else {
                    this.device = device

                    var times: number[] = device.getTimes()
                    var timeAverage: number = 0
                    for (let time of times) {
                        timeAverage += time
                    }
                    timeAverage /= times.length
                    timeAverage *= 100
                    var seconds: number = Math.floor(timeAverage % 60)
                    var minutes: number = Math.floor(timeAverage / 60) % 60
                    var hours: number = Math.floor(timeAverage / 60 / 60)
                    var potentialsot: string = (hours > 9 ? hours : "0" + hours) + ":"
                    potentialsot += (minutes > 9 ? minutes : "0" + minutes) + ":"
                    potentialsot += seconds > 9 ? seconds : "0" + seconds

                    var deviceSettings = device.getCommands()
                    var settings: string = ""
                    if (deviceSettings.length > 0) {
                        for (let i = 0; i < deviceSettings.length; i++) {
                            if (i != 0) settings += "<br><br>"
                            settings += deviceSettings[i]
                        }
                    } else {
                        settings = "Default (no modifications made in the app)"
                    }

                    this.contentList = [
                        ["Android Version", device.getAndroidVersion()],
                        ["Kernel Version", device.getKernelVersion()],
                        ["Board", device.getBoard()],
                        ["Fingerprint", device.getFingerprint()],
                        ["Potential SOT", potentialsot + " (when screen is always on)"],
                        ["CPU Score", device.getCpu() + " (lower is better)"],
                        ["CPU Information", device.getCpuInfo()],
                        ["Settings", settings]
                    ]
                }
            })
        })
    }

    newLine(index: number) {
        return index == 0 ? "" : "<br>"
    }

}