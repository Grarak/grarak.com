import { Component } from '@angular/core'
import { Router, ActivatedRoute, Params } from '@angular/router'

import { KernelAdiutorService, KernelAdiutorDevice } from '../services/kerneladiutor.service'

@Component({
    selector: `kerneladiutor-page`,
    template: `
        <pageparent-view>
            <div style="margin-bottom:20px;">
                <button md-raised-button color="accent" (click)="onPrevious()" style="margin-left:20px" [style.visibility]="showPrevious?'visible':'hidden'">Previous</button>
                <button md-raised-button color="accent" (click)="onNext()" style="float:right;margin-right:20px" [style.visibility]="showNext?'visible':'hidden'">Next</button>
            </div>
            <div style="margin-top:10px;margin-bottom:10px" *ngFor="let device of currentDevices; let i = index">
                <card-view>
                    <card-title>
                        <span style="font-size:smaller">{{(page - 1)*10+i+1}}. {{device.getVendor()}} {{device.getModel()}}</span>
                    </card-title>
                    <card-content>
                        {{device.getBoard()}}
                    </card-content>
                </card-view>
            </div>
        </pageparent-view>
    `
})
export class KernelAdiutorComponent {

    constructor(
        private kaService: KernelAdiutorService,
        private router: Router,
        private route: ActivatedRoute
    ) { }

    page: number
    currentDevices: KernelAdiutorDevice[]

    showPrevious: boolean
    showNext: boolean

    ngOnInit() {
        this.route.params.forEach((params: Params) => {

            this.currentDevices = []

            this.page = Number(params['page'])
            if (this.page > 0) {
                this.showPrevious = this.page > 1
                this.showNext = false
                this.kaService.getDevices(this.page).forEach((device: KernelAdiutorDevice) => {
                    if (device == null) {
                        this.router.navigate(['404'])
                    } else {
                        this.currentDevices.push(device)
                    }
                })

                this.kaService.getDevices(this.page + 1).forEach((device: KernelAdiutorDevice) => {
                    if (device != null) {
                        this.showNext = true
                    }
                })
            } else {
                this.router.navigate(['404'])
            }
        })
    }

    onPrevious() {
        this.router.navigate(['kerneladiutor/page/' + (this.page - 1)])
    }

    onNext() {
        this.router.navigate(['kerneladiutor/page/' + (this.page + 1)])
    }

}
