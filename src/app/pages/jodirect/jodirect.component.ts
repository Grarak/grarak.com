import {Component} from '@angular/core'
import {Router} from '@angular/router'

@Component({
    selector: `jodirect-page`,
    template: `
        <div id="parent">
            <button md-raised-button color="accent" class="button" (click)="onGenTokenClick()">Generate token</button>
            <br>
            <button md-raised-button color="accent" class="button" (click)="onSendClick()">Send message</button>
            <br>
            <button md-raised-button color="accent" class="button" (click)="onLoginClick()">Login</button>
            <br>
        </div>
    `,
    styles: [`
        #parent {
            position: absolute;
            transform: translateX(-50%) translateY(-50%);
            left: 50%;
            top: 50%;
            text-align: center;
        }

        .button {
            margin-top: 10px;
            margin-bottom: 10px;
        }
    `]
})
export class JoDirectComponent {

    constructor(private router: Router) {
    }

    onGenTokenClick() {
        this.router.navigate(['jodirect/gentoken'])
    }

    onSendClick() {
        this.router.navigate(['jodirect/send'])
    }

    onLoginClick() {
        this.router.navigate(['jodirect/login'])
    }

}
