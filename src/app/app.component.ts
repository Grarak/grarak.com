import {Component, ViewChild, HostListener} from '@angular/core'
import {Router, NavigationStart} from '@angular/router'

import {NavDrawerComponent} from './views/navdrawer.component'

import {Utils} from './utils/utils'

@Component({
    selector: 'app-view',
    template: `
        <div>
            <navdrawer-view #navdrawer [navbarOpened]="navbarOpened">
                <navbar-content>
                    <md-list>
                        <md-list-item *ngFor="let item of navbarItems">
                            <a routerLink="{{item.route}}" routerLinkActive="selected"
                               [routerLinkActiveOptions]="item.options" (click)="onNavbarItemClicked(item)">
                                <p style="margin-top: 0">{{item.title}}</p>
                            </a>
                        </md-list-item>
                    </md-list>
                </navbar-content>

                <navbar-page-content>
                    <toolbar-view [title]="toolbarTitle" (menuClicked)="onToolbarMenuClicked()"
                                  [menuDisplay]="menuDisplay">
                        Page
                    </toolbar-view>

                    <router-outlet></router-outlet>
                </navbar-page-content>
            </navdrawer-view>
        </div>
    `,
    styles: [
            `
            a {
                color: rgba(0, 0, 0, .6);
                text-decoration: none;
            }

            a:hover {
                color: black;
            }

            a:visited {
                color: rgba(0, 0, 0, .7);
            }

            .selected {
                color: black;
                font-weight: bold;
            }
        `
    ]
})
export class AppComponent {

    navbarItems: any[] = [
        {route: '/', title: 'About me', options: {exact: true}},
        {route: '/kerneladiutor', title: 'Kernel Adiutor', options: {exact: false}},
        {route: '/jodirect', title: 'JoDirect', options: {exact: false}}
    ]

    profile_pic: string = Utils.getAsset('profile_pic.jpg')
    ic_github: string = Utils.getAsset('ic_github.svg')

    @ViewChild('navdrawer') navdrawer: NavDrawerComponent

    navbarOpened: boolean
    menuDisplay: string
    toolbarTitle: string

    constructor(private router: Router) {
    }

    ngOnInit() {
        this.onWindowResize(window.innerWidth)
        this.router.events.subscribe((data) => {
            this.toolbarTitle = null
            for (const item of this.navbarItems) {
                const route = <NavigationStart>data
                if ((item.options.exact && route.url === item.route)
                    || (!item.options.exact && route.url.startsWith(item.route))) {
                    this.toolbarTitle = item.title
                }
            }
            if (this.toolbarTitle == null) {
                this.toolbarTitle = 'Not found'
            }
        })
    }

    @HostListener('window:resize', ['$event'])
    onResize(event) {
        this.onWindowResize(event.target.innerWidth)
    }

    onWindowResize(size: number) {
        this.menuDisplay = size > 700 ? 'none' : 'block'
    }

    onToolbarMenuClicked() {
        this.navdrawer.toggle()
    }

    onNavbarItemClicked(item: any) {
        this.navdrawer.toggle()
        this.toolbarTitle = item.title
    }

}
