import { Component, HostListener } from '@angular/core'

@Component({
    selector: `pageparent-view`,
    template: `
        <div [style.margin-left]="parentMargin" [style.margin-right]="parentMargin" style="margin-top:20px;margin-bottom:20px">
            <ng-content></ng-content>
        </div>
    `
})
export class PageParentComponent {

    parentMargin: string

    ngOnInit() {
        this.onWindowResize(window.innerWidth)
    }

    @HostListener('window:resize', ['$event'])
    onResize(event) {
        this.onWindowResize(event.target.innerWidth)
    }

    onWindowResize(size: number) {
        this.parentMargin = size > 1200 ? (size * 0.125) + "px" : size > 480 ? "30px" : "0"
    }

}
