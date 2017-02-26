import { Component } from '@angular/core'
import { ActivatedRoute, Params } from '@angular/router'

@Component({
    selector: `kerneladiutor-deviceinfo-page`,
    template: `
        {{id}}
    `
})
export class KernelAdiutorDeviceInfoComponent {

    id: string

    constructor(
        private route: ActivatedRoute
    ) { }

    ngOnInit() {
        this.route.params.forEach((params: Params) => {
            
            this.id = params['id']

        })
    }

}