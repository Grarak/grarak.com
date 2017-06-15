import {Component} from '@angular/core'
import {Router} from '@angular/router'

@Component({
    selector: `jodirect-login-page`,
    template: `
        <div id="parent">
            <md-input-container>
                <input mdInput placeholder="Token" [(ngModel)]="token">
            </md-input-container>
            <md-input-container>
                <input mdInput type="password" placeholder="Password" [(ngModel)]="password">
            </md-input-container>
            <span [style.display]="errorDisplay" style="color:red;font-style: italic;font-size: 14px">{{error}}</span>
            <br>
            <button md-raised-button color="accent" class="button" (click)="onLogin()">Login</button>
        </div>
    `,
    styles: [
            `
            #parent {
                position: absolute;
                top: 50%;
                left: 50%;
                transform: translateX(-50%) translateY(-50%);
                text-align: center;
            }
        `
    ]
})
export class JoDirectLoginComponent {

    token = ''
    password = ''
    error = ''

    errorDisplay = 'none'

    constructor(private router: Router) {
    }

    onLogin() {
        if (this.token === '') {
            this.error = 'Token is empty!'
        } else if (this.password === '') {
            this.error = 'Password is empty!'
        }

        if (this.error !== '') {
            this.errorDisplay = 'block'
        } else {
            this.router.navigate(['jodirect/messages'], {queryParams: {token: this.token, password: this.password}})
        }
    }

}
