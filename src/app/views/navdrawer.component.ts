import {Component, NgZone, ViewChild, HostListener} from '@angular/core'
import {MdSidenav} from '@angular/material'

import {Utils} from '../utils/utils'

@Component({
    selector: `navdrawer-view`,
    template: `
        <md-sidenav-container>

            <md-sidenav #sidenav [mode]="containerMode" [style.width]="navbarWidth"
                        [opened]="windowSize > 700 || navbarOpened">

                <img [src]="profile_pic" [style.width]="navbarWidth">

                <div id="navbar-content">
                    <ng-content select="navbar-content"></ng-content>
                </div>

            </md-sidenav>

            <div class="md-sidenav-content">
                <ng-content select="navbar-page-content"></ng-content>
            </div>

        </md-sidenav-container>
    `,
    styles: [
            `
            md-sidenav-container {
                position: absolute;
                height: 100%;
                width: 100%;
            }

            #navbar-content {
                padding: 16px;
            }
        `
    ],
    inputs: ['navbarOpened']
})
export class NavDrawerComponent {

    @ViewChild('sidenav') sideNav: MdSidenav

    profile_pic: string = Utils.getAsset('profile_pic.jpg')

    containerMode: string
    navbarWidth: string
    navbarOpened: boolean
    windowSize: number

    ngOnInit() {
        this.onWindowResize(window.innerWidth)
    }

    @HostListener('window:resize', ['$event'])
    onResize(event) {
        this.onWindowResize(event.target.innerWidth)
    }

    onWindowResize(size: number) {
        this.windowSize = size
        if (size > 700) {
            this.containerMode = 'side'
            this.navbarWidth = '250px'
            this.navbarOpened = true
            if (this.sideNav != null && this.sideNav._isClosed) {
                this.sideNav.toggle()
            }
        } else {
            this.containerMode = 'over'
            this.navbarWidth = (size <= 480 ? size - (size * 0.17) : size / 2) + "px"
        }
    }

    toggle() {
        if (this.windowSize <= 700) {
            this.sideNav.toggle()
        }
    }

}
