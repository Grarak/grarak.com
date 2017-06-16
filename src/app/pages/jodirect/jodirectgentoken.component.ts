import {Component} from '@angular/core'
import {Router} from '@angular/router'
import {Observable} from 'rxjs/Rx'

import {JoDirectService, JoDirectToken} from '../../services/jodirect.service'
import {Utils} from '../../utils/utils'

@Component({
    selector: `jodirect-gen-token-page`,
    template: `
        <div id="parent">
            <span [style.display]="tokenDisplay">Token: {{token}}</span>
            <span [style.display]="tokenDisplay">Password: {{password}}</span><br>
            <span [style.display]="tokenDisplay">Please keep in mind that you won't be able to generate a new token
                in the next 5 minutes, therefore write down your current token and password.</span>
            <span [style.display]="tokenDisplay"><br><button md-raised-button color="accent"
                                                             class="button"
                                                             (click)="onLogin()">Login with current token</button></span>
            <span [style.display]="statusDisplay">{{status}}</span>
            <md-spinner [style.display]="loadingDisplay"
                        style="width: 3em;left:50%;position: absolute;transform: translateX(-50%)"></md-spinner>
        </div>
    `,
    styles: [
            `
            #parent {
                position: absolute;
                top: 50%;
                left: 50%;
                transform: translateX(-50%) translateY(-50%);
            }
        `
    ]
})
export class JoDirectGenTokenComponent {

    token: string
    password: string
    status: string

    tokenDisplay = 'none'
    statusDisplay = 'none'
    loadingDisplay = 'block'

    timeleft: number
    timer: any

    constructor(private router: Router,
                private joDirectService: JoDirectService) {
    }

    ngOnInit() {
        this.joDirectService.genToken().forEach((token: JoDirectToken) => {
            this.loadingDisplay = 'none'
            if (token === null) {
                this.statusDisplay = 'block'
                this.status = 'Something went wrong!'
            } else if (!token.success()) {
                this.statusDisplay = 'block'
                this.timeleft = Math.ceil(token.getTimeleft())
                this.updateStatus()

                this.timer = Observable.timer(0, 1000).subscribe(() => {
                    this.timeleft--
                    if (this.timeleft >= 0) {
                        this.updateStatus()
                    }
                })
            } else {
                this.tokenDisplay = 'block'
                this.token = token.getToken()
                this.password = token.getPassword()
            }
        })
    }

    updateStatus() {
        this.status = 'You already generated a token. Wait '
            + Utils.formatSeconds(this.timeleft) + ' to generate your next token.'
    }

    ngOnDestroy() {
        if (this.timer != null) {
            this.timer.unsubscribe()
        }
    }

    onLogin() {
        this.router.navigate(['jodirect/messages'], {queryParams: {token: this.token, password: this.password}})
    }

}
