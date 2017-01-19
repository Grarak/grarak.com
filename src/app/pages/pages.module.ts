import { BrowserModule } from '@angular/platform-browser'
import { NgModule } from '@angular/core'
import { RouterModule } from '@angular/router'

import { ViewsModule } from '../views/views.module'

import { AboutMeComponent } from './aboutme.component'
import { KernelAdiutorComponent } from './kerneladiutor.component'
import { NotFoundComponent } from './notfound.component'
import { PageParentComponent } from './pageparent.component'

@NgModule({
    imports: [
        BrowserModule,
        ViewsModule,
        RouterModule.forRoot([
            { path: '', component: AboutMeComponent },
            { path: 'kerneladiutor', component: KernelAdiutorComponent },
            { path: '**', component: NotFoundComponent },
        ], {})
    ],
    declarations: [
        AboutMeComponent,
        KernelAdiutorComponent,
        NotFoundComponent,
        PageParentComponent
    ],
    exports: [
        RouterModule
    ]
})
export class PagesModule {
}
