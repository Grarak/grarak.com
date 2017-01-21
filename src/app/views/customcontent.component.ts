import { Component } from '@angular/core'

@Component({
    selector: `card-title,card-content,navbar-content,navbar-page-content`,
    template: `
        <ng-content></ng-content>
    `
})
export class CustomComponent { }
