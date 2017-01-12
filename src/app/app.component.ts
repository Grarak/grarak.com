import { Component, NgZone } from '@angular/core'

import { Utils } from './utils/utils'

@Component({
    selector: 'app-view',
    template: `
        <div #parent>
            <toolbar-view title="Home">
                <div id="toolbar_parent" align="center">
                    <span id="toolbar_content_text">Welcome!</span>
                </div>
            </toolbar-view>
            <div style="width: 100%">
                <card-view id="card" [style.width.]="cardwitdh">
                    <div content>
                        <p align="center">Hi welcome to my website.</p>
                        <p align="center">My name is Willi Ye and I am a hobby programmer.</p>
                         <div align="center">
                            <a target="_blank" href="https://www.github.com/Grarak">
                                <img [src]="ic_github" width="25" height="25">
                            </a>
                        </div>
                    </div>
                </card-view>
                <img id="profile_pic" [src]="profile_pic">
            </div>
        </div>
    `,
    styles: [
        `
            #toolbar_parent {
                position: absolute;
                height: 100%;
                width: 100%;
            }
            #toolbar_content_text {
                font-size: 2em;
                position: absolute;
                top: 40%;
                left: 50%;
                transform: translateX(-50%);
                color: white;
                font-family: sans-serif;
            }
            #profile_pic {
                position: absolute;
                width: 8em;
                height: auto;
                border-radius: 50%;
                margin-top: -4em;
                margin-left: 50%;
                margin-right: 50%;
                transform: translateX(-50%);
            }
            #card {
                position: absolute;
                margin-top: 0.5em;
                left: 50%;
                transform: translateX(-50%);
            }
        `
    ]
})
export class AppComponent {

    profile_pic: string = Utils.getAsset('profile_pic.jpg')
    ic_github: string = Utils.getAsset('ic_github.svg')

    cardwitdh: string

    constructor(ngZone: NgZone) {
        this.cardwitdh = (window.innerWidth - window.innerWidth / 12) + "px"
        window.onresize = () => {
            ngZone.run(() => {
                this.cardwitdh = (window.innerWidth - window.innerWidth / 12) + "px"
            })
        }
    }

}
