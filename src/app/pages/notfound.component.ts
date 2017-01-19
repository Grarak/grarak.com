import { Component, HostListener } from '@angular/core'

@Component({
    selector: `notfound-page`,
    template: `
        <pageparent-view>
            <card-view>
                <card-content>
                    <span style="color:rgba(0,0,0,.7);font-size:larger;position:absolute;left:50%;transform:translateX(-50%) translateY(-50%)">
                        Couldn't find your requested page!
                    </span>
                </card-content>
            </card-view>
        </pageparent-view>
    `
})
export class NotFoundComponent {
}
