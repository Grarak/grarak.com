import {NgModule} from '@angular/core'

import {KernelAdiutorService} from './kerneladiutor.service'
import {JoDirectService} from './jodirect.service'

@NgModule({
    providers: [
        KernelAdiutorService,
        JoDirectService
    ]
})
export class ServicesModule {
}
