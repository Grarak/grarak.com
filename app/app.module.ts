import { NgModule } from '@angular/core'
import { BrowserModule } from '@angular/platform-browser'

import { ViewModule } from './views/view.module'

import { AppComponent } from './app.component'

@NgModule({
    imports: [
        BrowserModule,
        ViewModule
    ],
    declarations: [
        AppComponent
    ],
    bootstrap: [AppComponent]
})
export class AppModule {
}
