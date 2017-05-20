import 'hammerjs';

import {BrowserModule} from '@angular/platform-browser'
import {NgModule} from '@angular/core'
import {FormsModule} from '@angular/forms'
import {HttpModule} from '@angular/http'

import {AppComponent} from './app.component'

import {PagesModule} from './pages/pages.module'
import {ViewsModule} from './views/views.module'

@NgModule({
    declarations: [
        AppComponent
    ],
    imports: [
        BrowserModule,
        FormsModule,
        HttpModule,
        PagesModule,
        ViewsModule
    ],
    providers: [],
    bootstrap: [AppComponent]
})
export class AppModule {
}
