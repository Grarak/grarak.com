import {Component, EventEmitter, Output} from '@angular/core'

import {Utils} from '../utils/utils'

@Component({
    selector: `toolbar-view`,
    template: `
        <div>
            <md-toolbar class="shadow" color="primary">
                <div [style.display]="menuDisplay" style="height: 100%">
                    <md-icon id="hamburger" (click)="onMenuClick()">menu</md-icon>
                </div>
                <span>{{title}}</span>
            </md-toolbar>
        </div>
    `,
    styles: [
            `
            .shadow {
                box-shadow: 0 4px 8px 0 rgba(0, 0, 0, 0.2), 0 6px 20px 0 rgba(0, 0, 0, 0.19);
            }

            #hamburger {
                margin-right: 20px;
                top: 50%;
                transform: translateY(-50%);
                position: relative;
                cursor: pointer;
            }
        `
    ],
    inputs: ['title', 'menuDisplay'],
})
export class ToolbarComponent {

    title: string
    menuDisplay = 'none'

    @Output() menuClicked = new EventEmitter()

    onMenuClick() {
        this.menuClicked.emit()
    }

}
