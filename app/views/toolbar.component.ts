import { Component } from '@angular/core';

@Component({
    selector: `toolbar-view`,
    template: `
        <div>
            <md-toolbar class="shadow" color="primary">
                <span>{{title}}</span>
            </md-toolbar>
            <div id="bg">
                <ng-content></ng-content>
            </div>
        </div>
    `,
    styles: [
        `
            .shadow {
                box-shadow: 0 4px 8px 0 rgba(0, 0, 0, 0.2), 0 6px 20px 0 rgba(0, 0, 0, 0.19);
            }

            #bg {
                position: relative;
                background-color: #2a7289;
                width: 100%;
                min-height: 20em;
                z-index: -1;
            }
        `
    ],
    inputs: ['title', 'fixed']
})
export class ToolbarComponent {

    title: string
    fixed: boolean

}