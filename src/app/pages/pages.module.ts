import { BrowserModule } from '@angular/platform-browser'
import { NgModule } from '@angular/core'
import { RouterModule } from '@angular/router'

import { ServicesModule } from '../services/services.module'
import { ViewsModule } from '../views/views.module'

import { AboutMeComponent } from './aboutme.component'
import { KernelAdiutorComponent } from './kerneladiutor/kerneladiutor.component'
import { KernelAdiutorDeviceInfoComponent } from './kerneladiutor/kerneladiutordeviceinfo.component'
import { NotFoundComponent } from './notfound.component'
import { PageParentComponent } from './pageparent.component'

@NgModule({
    imports: [
        BrowserModule,
        RouterModule.forRoot([
            { path: '', component: AboutMeComponent },
            { path: 'kerneladiutor/page/:page', component: KernelAdiutorComponent },
            { path: 'kerneladiutor/id/:id', component: KernelAdiutorDeviceInfoComponent },
            { path: 'kerneladiutor', redirectTo: 'kerneladiutor/page/1' },
            { path: '404', component: NotFoundComponent },
            { path: '**', redirectTo: '404'  },
        ], {}),
        ServicesModule,
        ViewsModule
    ],
    declarations: [
        AboutMeComponent,
        KernelAdiutorComponent,
        KernelAdiutorDeviceInfoComponent,
        NotFoundComponent,
        PageParentComponent
    ],
    exports: [
        RouterModule
    ]
})
export class PagesModule {
}
