import {BrowserModule} from '@angular/platform-browser'
import {NgModule} from '@angular/core'
import {RouterModule} from '@angular/router'
import {FormsModule} from '@angular/forms'

import {ServicesModule} from '../services/services.module'
import {ViewsModule} from '../views/views.module'

import {AboutMeComponent} from './aboutme.component'
import {KernelAdiutorComponent} from './kerneladiutor/kerneladiutor.component'
import {KernelAdiutorDeviceInfoComponent} from './kerneladiutor/kerneladiutordeviceinfo.component'

import {JoDirectComponent} from './jodirect/jodirect.component'
import {JoDirectGenTokenComponent} from './jodirect/jodirectgentoken.component'
import {JoDirectLoginComponent} from './jodirect/jodirectlogin.component'
import {JoDirectMessagesComponent} from './jodirect/jodirectmessages.component'
import {JoDirectSendComponent} from './jodirect/jodirectsend.component'

import {NotFoundComponent} from './notfound.component'
import {PageParentComponent} from './pageparent.component'

@NgModule({
    imports: [
        BrowserModule,
        FormsModule,
        RouterModule.forRoot([
            {path: '', component: AboutMeComponent},
            {path: 'kerneladiutor/page/:page', component: KernelAdiutorComponent},
            {path: 'kerneladiutor/id/:id', component: KernelAdiutorDeviceInfoComponent},
            {path: 'kerneladiutor', redirectTo: 'kerneladiutor/page/1'},
            {path: 'jodirect', component: JoDirectComponent},
            {path: 'jodirect/gentoken', component: JoDirectGenTokenComponent},
            {path: 'jodirect/login', component: JoDirectLoginComponent},
            {path: 'jodirect/messages', component: JoDirectMessagesComponent},
            {path: 'jodirect/send', component: JoDirectSendComponent},
            {path: '404', component: NotFoundComponent},
            {path: '**', redirectTo: '404'},
        ], {}),
        ServicesModule,
        ViewsModule
    ],
    declarations: [
        AboutMeComponent,
        KernelAdiutorComponent,
        KernelAdiutorDeviceInfoComponent,
        JoDirectComponent,
        JoDirectGenTokenComponent,
        JoDirectLoginComponent,
        JoDirectMessagesComponent,
        JoDirectSendComponent,
        NotFoundComponent,
        PageParentComponent
    ],
    exports: [
        RouterModule
    ]
})
export class PagesModule {
}
