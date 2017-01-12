import { NgModule } from '@angular/core'

import { MaterialModule } from '@angular/material'

import { CardComponent } from './card.component'
import { ToolbarComponent } from './toolbar.component'

@NgModule({
    imports: [
        MaterialModule.forRoot()
    ],
    exports: [
        CardComponent,
        ToolbarComponent,
    ],
    declarations: [
        CardComponent,
        ToolbarComponent
    ]
})
export class ViewsModule {
}
