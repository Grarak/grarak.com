import { NgModule } from '@angular/core'

import { MaterialModule } from '@angular/material'

import { CardComponent } from './card.component'
import { CustomComponent } from './customcontent.component'
import { NavDrawerComponent } from './navdrawer.component'
import { ToolbarComponent } from './toolbar.component'

@NgModule({
    imports: [
        MaterialModule.forRoot()
    ],
    exports: [
        MaterialModule,
        CardComponent,
        CustomComponent,
        NavDrawerComponent,
        ToolbarComponent,
    ],
    declarations: [
        CardComponent,
        CustomComponent,
        NavDrawerComponent,
        ToolbarComponent
    ]
})
export class ViewsModule {
}
